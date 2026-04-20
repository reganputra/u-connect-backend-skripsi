package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/routes"
	"github.com/reganputra/skripsi-backend/scheduler"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
	"github.com/reganputra/skripsi-backend/ws"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// Redirect Go's standard logger to stdout so that [WS/*] structured logs
	// appear on the same stream as Fiber's access log. Without this, log.Printf
	// writes to stderr and the two streams appear separately in most terminals.
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
	mentorRepo := repository.NewMentorRepository(db)
	mentorRequestRepo := repository.NewMentorRequestRepository(db)
	mentoringSessionRepo := repository.NewMentoringSessionRepository(db)
	followRepo := repository.NewFollowRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	// ── WebSocket Hub + Notification Service (must come first) ──────────────
	hub := ws.NewHub()
	go hub.Run()
	notifSvc := service.NewNotificationService(notifRepo, hub)

	// ── Services ──────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, profileRepo)
	profileSvc := service.NewProfileService(profileRepo)
	companySvc := service.NewCompanyService(companyRepo, userRepo)
	portfolioSvc := service.NewPortfolioService(portfolioRepo)
	feedSvc := service.NewFeedService(postRepo, commentRepo, reactionRepo, voteRepo, userRepo, notifSvc)
	groupSvc := service.NewGroupService(groupRepo, memberRepo, articleRepo, gCommentRepo, gReactionRepo, notifSvc)
	eventSvc := service.NewEventService(eventRepo, agendaRepo, regRepo)
	jobSvc := service.NewJobService(jobRepo, jobAppRepo, notifSvc)
	reportSvc := service.NewReportService(reportRepo)
	adminSvc := service.NewAdminService(adminRepo, reportRepo, categoryRepo, notifSvc)
	recommendSvc := service.NewRecommendationService(mentorRepo)
	mentorSvc := service.NewMentorService(profileRepo, mentorRepo, mentorRequestRepo, mentoringSessionRepo, recommendSvc, userRepo, notifSvc)
	followSvc := service.NewFollowService(followRepo, userRepo, notifSvc)
	messageSvc := service.NewMessageService(messageRepo, followRepo)

	// ── Controllers ───────────────────────────────────────────────────────────
	authCtrl := controllers.NewAuthController(authSvc)
	profileCtrl := controllers.NewProfileController(profileSvc)
	directoryCtrl := controllers.NewDirectoryController(profileSvc, portfolioSvc)
	companyCtrl := controllers.NewCompanyController(companySvc)
	portfolioCtrl := controllers.NewPortfolioController(portfolioSvc)
	feedCtrl := controllers.NewFeedController(feedSvc)
	groupCtrl := controllers.NewGroupController(groupSvc)
	eventCtrl := controllers.NewEventController(eventSvc)
	jobCtrl := controllers.NewJobController(jobSvc)
	reportCtrl := controllers.NewReportController(reportSvc)
	adminCtrl := controllers.NewAdminController(adminSvc)
	mentorCtrl := controllers.NewMentorController(mentorSvc)
	followCtrl := controllers.NewFollowController(followSvc)
	messageCtrl := controllers.NewMessageController(messageSvc)
	notifCtrl := controllers.NewNotificationController(notifSvc)

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
	routes.RegisterAuthRoutes(app, authCtrl)
	routes.RegisterProfileRoutes(app, profileCtrl)
	routes.SetupDirectoryRoutes(app, directoryCtrl)
	routes.RegisterCompanyRoutes(app, companyCtrl)
	routes.RegisterPortfolioRoutes(app, portfolioCtrl)
	routes.RegisterFeedRoutes(app, feedCtrl)
	routes.RegisterGroupRoutes(app, groupCtrl)
	routes.RegisterEventRoutes(app, eventCtrl)
	routes.RegisterJobRoutes(app, jobCtrl)
	routes.RegisterReportRoutes(app, reportCtrl)
	routes.RegisterAdminRoutes(app, adminCtrl)
	routes.RegisterMentorRoutes(app, mentorCtrl)
	routes.SetupFollowRoutes(app, followCtrl)
	routes.SetupMessageRoutes(app, messageCtrl, hub, messageSvc, userRepo, notifSvc)
	routes.SetupNotificationRoutes(app, notifCtrl)

	// ── Background Schedulers ───────────────────────────────────────────────────
	go scheduler.StartEventReminderScheduler(db, notifSvc)
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
