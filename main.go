package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/container"
	"github.com/reganputra/skripsi-backend/routes"
	"github.com/reganputra/skripsi-backend/scheduler"
	"github.com/reganputra/skripsi-backend/utils"
	"github.com/reganputra/skripsi-backend/ws"
)

func main() {

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.Lmsgprefix) // match Fiber's time-only prefix style

	// Load environment variables from .env file
	utils.LoadEnvFile(".env")

	// Validate required environment variables before starting
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("❌ JWT_SECRET environment variable is not set")
	}

	// Connect to the database
	config.ConnectDB()

	// Initialize Cloudinary
	config.ConnectCloudinary()

	// ── C-1: Synchronous migration before serving traffic (no background goroutine).
	if err := config.RunMigrations(config.DB); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	db := config.DB

	// ── WebSocket Hub ──────────────
	hub := ws.NewHub()
	go hub.Run()

	// ── Build Dependencies via Container ──────────────────────────────────────
	c := container.Build(db, hub)

	// ── Fiber app ─────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Alumni Community Platform API v1.0",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// Panic recovery — catches any handler panic, logs it with stack
	// trace, and returns 500 without crashing the server.
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			log.Printf("[PANIC] recovered: %v | %s %s", e, c.Method(), c.Path())
		},
	}))

	app.Use(logger.New())

	// ── CORS ──────────────────────────────────────────────────────────────────
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" { // prod
		allowedOrigins = "http://localhost:5173" // dev
	}
	app.Use(cors.New(cors.Config{
		Next: func(c *fiber.Ctx) bool {
			return strings.HasPrefix(c.Path(), "/api/ws")
		},
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Extensions, Connection, Upgrade, ngrok-skip-browser-warning",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: allowedOrigins != "*", // credentials not allowed with wildcard
	}))

	// ── Health probes (C-1) ───────────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
		})
	})
	app.Get("/ready", func(c *fiber.Ctx) error {
		sqlDB, err := config.DB.DB()
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not_ready",
				"error":  "db handle unavailable",
			})
		}
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not_ready",
				"error":  "db ping failed",
			})
		}
		var exists bool
		if err := config.DB.WithContext(ctx).Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users')").Scan(&exists).Error; err != nil || !exists {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not_ready",
				"error":  "schema not ready",
			})
		}
		return c.JSON(fiber.Map{
			"status":  "ready",
			"message": "Server is ready",
		})
	})

	// ── Routes ────────────────────────────────────────────────────────────────
	routes.RegisterAuthRoutes(app, c.AuthCtrl)
	routes.RegisterProfileRoutes(app, c.ProfileCtrl)
	routes.SetupDirectoryRoutes(app, c.DirectoryCtrl)
	routes.RegisterCompanyRoutes(app, c.CompanyCtrl)
	routes.RegisterPortfolioRoutes(app, c.PortfolioCtrl)
	routes.RegisterFeedRoutes(app, c.FeedCtrl, db)
	routes.RegisterGroupRoutes(app, c.GroupCtrl, db)
	routes.RegisterEventRoutes(app, c.EventCtrl, db)
	routes.RegisterJobRoutes(app, c.JobCtrl, db)
	routes.RegisterReportRoutes(app, c.ReportCtrl)
	routes.RegisterAdminRoutes(app, c.AdminCtrl, c.AuthCtrl, c.EvalCtrl)
	routes.RegisterMentorRoutes(app, c.MentorCtrl)
	routes.SetupFollowRoutes(app, c.FollowCtrl)
	routes.SetupMessageRoutes(app, c.MessageCtrl, hub, c.MessageSvc, c.UserRepo, c.NotifSvc)
	routes.SetupNotificationRoutes(app, c.NotifCtrl)
	routes.RegisterActivityRoutes(app, c.ActivityCtrl)
	routes.RegisterAnalyticsRoutes(app, c.AnalyticsCtrl)
	routes.RegisterAnnouncementRoutes(app, c.AnnouncementCtrl)

	// ── Background Schedulers (with context for graceful shutdown) ──────────────
	schedulerCtx, cancelSchedulers := context.WithCancel(context.Background())
	go scheduler.StartEventReminderScheduler(schedulerCtx, db, c.NotifSvc)
	go scheduler.StartEventStatusScheduler(schedulerCtx, db)
	// go scheduler.StartAnalyticsSnapshotScheduler(schedulerCtx, c.AnalyticsCtrl.Svc())
	// Backfill: compute snapshots for last 90 days on first boot
	// go func() {
	// 	if err := c.AnalyticsCtrl.Svc().BackfillSnapshots(90); err != nil {
	// 		log.Printf("[analytics] backfill error: %v", err)
	// 	}
	// }()

	// ── Start server ──────────────────────────────────────────────────────────
	// Prefer the `PORT` env var (platforms like Vercel provide this),
	// fall back to `APP_PORT`, then to `8080` for local dev.
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdownChan
		log.Println("⏳ Shutting down gracefully...")

		// 1. Stop accepting new HTTP requests (30 second drain window)
		if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
			log.Printf("⚠️  App shutdown error: %v", err)
		}

		// 2. Cancel context → stops schedulers
		cancelSchedulers()

		// 3. Close WebSocket hub (signals all clients to disconnect)
		hub.Close()

		// 4. Close database connection pool
		if sqlDB, err := config.DB.DB(); err == nil {
			_ = sqlDB.Close()
		}

		log.Println("✅ Server stopped cleanly")
	}()

	log.Printf("🚀 Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil && !errors.Is(err, net.ErrClosed) {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
