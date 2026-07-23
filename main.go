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
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/routes"
	"github.com/reganputra/skripsi-backend/scheduler"
	"github.com/reganputra/skripsi-backend/utils"
	"github.com/reganputra/skripsi-backend/ws"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

	// ── Background: migrate + seed (runs after HTTP server is already listening)
	// instead of timing out while AutoMigrate runs 20+ slow SQL queries.
	go func() {
		if err := config.DB.AutoMigrate(
			&models.User{},
			&models.UserProfile{},
			&models.UserExperience{},
			&models.CompanyProfile{},
			&models.PortfolioItem{},
			&models.Post{},
			&models.PostImage{},
			&models.Comment{},
			&models.Reaction{},
			&models.Vote{},
			&models.Group{},
			&models.GroupMember{},
			&models.GroupArticle{},
			&models.GroupArticleImage{},
			&models.GroupComment{},
			&models.GroupReaction{},
			&models.Event{},
			&models.EventAgenda{},
			&models.EventRegistration{},
			&models.Job{},
			&models.JobApplication{},
			&models.Report{},
			&models.Category{},
			&models.MentorRequest{},
			&models.MentoringSession{},
			&models.Follow{},
			&models.Message{},
			&models.Notification{},
			&models.PageView{},
			&models.DailyAnalyticsSnapshot{},
			&models.AdminActivityLog{},
			&models.Announcement{},
		); err != nil {
			log.Printf("❌ AutoMigrate failed: %v", err)
			return
		}
		log.Println("✅ Database migrated successfully!")

		if err := cleanupDuplicateFollows(config.DB); err != nil {
			log.Printf("❌ Failed to clean duplicate follows: %v", err)
		}

		seedGeneralCategory(config.DB)
		seedAdmin(config.DB)
	}()

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

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
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

func cleanupDuplicateFollows(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`
			DELETE FROM follows f
			USING (
				SELECT id,
					ROW_NUMBER() OVER (
						PARTITION BY follower_id, following_id
						ORDER BY (deleted_at IS NOT NULL) ASC, created_at DESC, id DESC
					) AS rn
				FROM follows
			) ranked
			WHERE f.id = ranked.id
				AND ranked.rn > 1
		`).Error
	})
}

// seedAdmin creates the first admin account from environment variables
// if no admin user already exists. Safe to call on every startup.
func seedAdmin(db *gorm.DB) {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	name := os.Getenv("ADMIN_NAME")

	if email == "" || password == "" {
		log.Println("⚠️  ADMIN_EMAIL or ADMIN_PASSWORD not set — skipping admin seeder")
		return
	}

	var count int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		log.Println("ℹ️  Admin user already exists — skipping seeder")
		return
	}

	if name == "" {
		name = "Administrator"
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ Failed to hash admin password: %v", err)
		return
	}

	admin := models.User{
		Name:     name,
		Email:    email,
		Password: string(hashed),
		Role:     "admin",
		IsActive: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("❌ Failed to seed admin user: %v", err)
		return
	}

	log.Printf("✅ Admin user seeded: %s", email)
}

// seedGeneralCategory ensures the fallback category exists for content cleanup.
func seedGeneralCategory(db *gorm.DB) {
	var count int64
	db.Model(&models.Category{}).Where("name = ?", "General").Count(&count)
	if count > 0 {
		return
	}

	if err := db.Create(&models.Category{Name: "General"}).Error; err != nil {
		log.Printf("❌ Failed to seed General category: %v", err)
		return
	}

	log.Println("✅ General category seeded")
}
