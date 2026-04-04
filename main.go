package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/routes"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

func main() {
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

	// Auto-migrate models
	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.UserExperience{},
		&models.CompanyProfile{},
		&models.PortfolioItem{},
		&models.Post{},
		&models.Comment{},
		&models.Reaction{},
		&models.Vote{},
		&models.Group{},
		&models.GroupMember{},
		&models.GroupArticle{},
		&models.GroupComment{},
		&models.GroupReaction{},
		&models.Event{},
		&models.EventAgenda{},
		&models.EventRegistration{},
		&models.Job{},
		&models.JobApplication{},
		&models.Report{},
		&models.Category{},
	); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}
	log.Println("✅ Database migrated successfully!")

	// ── Admin Seeder ──────────────────────────────────────────────────────────
	seedAdmin(config.DB)

	db := config.DB

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	profileRepo := repository.NewProfileRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	portfolioRepo := repository.NewPortfolioRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	voteRepo := repository.NewVoteRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	memberRepo := repository.NewGroupMemberRepository(db)
	articleRepo := repository.NewGroupArticleRepository(db)
	gCommentRepo := repository.NewGroupCommentRepository(db)
	gReactionRepo := repository.NewGroupReactionRepository(db)
	eventRepo := repository.NewEventRepository(db)
	agendaRepo := repository.NewEventAgendaRepository(db)
	regRepo := repository.NewEventRegistrationRepository(db)
	jobRepo := repository.NewJobRepository(db)
	jobAppRepo := repository.NewJobApplicationRepository(db)
	reportRepo := repository.NewReportRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo)
	profileSvc := service.NewProfileService(profileRepo)
	companySvc := service.NewCompanyService(companyRepo, userRepo)
	portfolioSvc := service.NewPortfolioService(portfolioRepo)
	feedSvc := service.NewFeedService(postRepo, commentRepo, reactionRepo, voteRepo)
	groupSvc := service.NewGroupService(groupRepo, memberRepo, articleRepo, gCommentRepo, gReactionRepo)
	eventSvc := service.NewEventService(eventRepo, agendaRepo, regRepo)
	jobSvc := service.NewJobService(jobRepo, jobAppRepo)
	reportSvc := service.NewReportService(reportRepo)
	adminSvc := service.NewAdminService(adminRepo, reportRepo, categoryRepo)

	// ── Controllers ───────────────────────────────────────────────────────────
	authCtrl := controllers.NewAuthController(authSvc)
	profileCtrl := controllers.NewProfileController(profileSvc)
	companyCtrl := controllers.NewCompanyController(companySvc)
	portfolioCtrl := controllers.NewPortfolioController(portfolioSvc)
	feedCtrl := controllers.NewFeedController(feedSvc)
	groupCtrl := controllers.NewGroupController(groupSvc)
	eventCtrl := controllers.NewEventController(eventSvc)
	jobCtrl := controllers.NewJobController(jobSvc)
	reportCtrl := controllers.NewReportController(reportSvc)
	adminCtrl := controllers.NewAdminController(adminSvc)

	// ── Fiber app ─────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName: "Alumni Community Platform API v1.0",
	})

	app.Use(logger.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// ── Routes ────────────────────────────────────────────────────────────────
	routes.RegisterAuthRoutes(app, authCtrl)
	routes.RegisterProfileRoutes(app, profileCtrl)
	routes.RegisterCompanyRoutes(app, companyCtrl)
	routes.RegisterPortfolioRoutes(app, portfolioCtrl)
	routes.RegisterFeedRoutes(app, feedCtrl)
	routes.RegisterGroupRoutes(app, groupCtrl)
	routes.RegisterEventRoutes(app, eventCtrl)
	routes.RegisterJobRoutes(app, jobCtrl)
	routes.RegisterReportRoutes(app, reportCtrl)
	routes.RegisterAdminRoutes(app, adminCtrl)

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
