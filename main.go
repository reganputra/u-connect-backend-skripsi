package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	if err := cleanupDuplicateFollows(config.DB); err != nil {
		log.Fatalf("❌ Failed to clean duplicate follows: %v", err)
	}

	// Initialize Cloudinary
	config.ConnectCloudinary()
	// Auto-migrate models
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
	); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}
	log.Println("✅ Database migrated successfully!")

	seedGeneralCategory(config.DB)

	// ── Admin Seeder ──────────────────────────────────────────────────────────
	seedAdmin(config.DB)

	db := config.DB

	// ── WebSocket Hub ──────────────
	hub := ws.NewHub()
	go hub.Run()

	// ── Build Dependencies via Container ──────────────────────────────────────
	c := container.Build(db, hub)

	// ── Fiber app ─────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName: "Alumni Community Platform API v1.0",
	})

	app.Use(logger.New())

	// ── CORS ──────────────────────────────────────────────────────────────────
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173" // dev default — restrict in production
	}
	app.Use(cors.New(cors.Config{
		Next: func(c *fiber.Ctx) bool {
			// Skip CORS for WebSocket upgrade requests — the WS auth
			// middleware handles these directly and CORS headers are
			// irrelevant once the connection is upgraded.
			return strings.HasPrefix(c.Path(), "/api/ws")
		},
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Sec-WebSocket-Key, Sec-WebSocket-Version, Sec-WebSocket-Extensions, Connection, Upgrade",
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
	routes.RegisterFeedRoutes(app, c.FeedCtrl)
	routes.RegisterGroupRoutes(app, c.GroupCtrl)
	routes.RegisterEventRoutes(app, c.EventCtrl)
	routes.RegisterJobRoutes(app, c.JobCtrl)
	routes.RegisterReportRoutes(app, c.ReportCtrl)
	routes.RegisterAdminRoutes(app, c.AdminCtrl)
	routes.RegisterMentorRoutes(app, c.MentorCtrl)
	routes.SetupFollowRoutes(app, c.FollowCtrl)
	routes.SetupMessageRoutes(app, c.MessageCtrl, hub, c.MessageSvc, c.UserRepo, c.NotifSvc)
	routes.SetupNotificationRoutes(app, c.NotifCtrl)
	routes.RegisterActivityRoutes(app, c.ActivityCtrl)

	// ── Background Schedulers ───────────────────────────────────────────────────
	go scheduler.StartEventReminderScheduler(db, c.NotifSvc)
	go scheduler.StartEventStatusScheduler(db)

	// ── Start server ──────────────────────────────────────────────────────────
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
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
