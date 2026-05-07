package container

import (
	"log"

	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/ws"
	"gorm.io/gorm"
)

type AppContainer struct {
	// Only exposing the repositories/services we need directly outside
	UserRepo   repository.UserRepository
	NotifSvc   service.NotificationService
	MessageSvc service.MessageService

	// All Controllers
	AuthCtrl      *controllers.AuthController
	ProfileCtrl   *controllers.ProfileController
	DirectoryCtrl *controllers.DirectoryController
	CompanyCtrl   *controllers.CompanyController
	PortfolioCtrl *controllers.PortfolioController
	FeedCtrl      *controllers.FeedController
	GroupCtrl     *controllers.GroupController
	EventCtrl     *controllers.EventController
	JobCtrl       *controllers.JobController
	ReportCtrl    *controllers.ReportController
	AdminCtrl     *controllers.AdminController
	MentorCtrl    *controllers.MentorController
	FollowCtrl    *controllers.FollowController
	MessageCtrl   *controllers.MessageController
	NotifCtrl     *controllers.NotificationController
	ActivityCtrl  *controllers.ActivityController
	AnalyticsCtrl *controllers.AnalyticsController
}

// Build wires up all dependencies and returns the container.
func Build(db *gorm.DB, hub *ws.Hub) *AppContainer {
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
	analyticsRepo := repository.NewAnalyticsRepository(db)

	if err := profileRepo.BackfillMissingPartnerProfiles(); err != nil {
		log.Fatalf("❌ Failed to backfill partner profiles: %v", err)
	}

	// ── Services ──────────────────────────────────────────────────────────────
	notifSvc := service.NewNotificationService(notifRepo, hub, db)
	authSvc := service.NewAuthService(userRepo, profileRepo)
	profileSvc := service.NewProfileService(profileRepo, userRepo)
	companySvc := service.NewCompanyService(companyRepo, userRepo, profileRepo)
	portfolioSvc := service.NewPortfolioService(portfolioRepo)
	feedSvc := service.NewFeedService(postRepo, commentRepo, reactionRepo, voteRepo, userRepo, notifSvc)
	groupSvc := service.NewGroupService(groupRepo, memberRepo, articleRepo, gCommentRepo, gReactionRepo, notifSvc)
	eventSvc := service.NewEventService(eventRepo, agendaRepo, regRepo)
	jobSvc := service.NewJobService(jobRepo, jobAppRepo, companyRepo, userRepo, notifSvc)
	reportSvc := service.NewReportService(reportRepo)
	adminSvc := service.NewAdminService(adminRepo, reportRepo, categoryRepo, notifSvc, db)
	recommendSvc := service.NewRecommendationService(mentorRepo)
	mentorSvc := service.NewMentorService(profileRepo, mentorRepo, mentorRequestRepo, mentoringSessionRepo, recommendSvc, userRepo, notifSvc)
	followSvc := service.NewFollowService(followRepo, userRepo, notifSvc)
	messageSvc := service.NewMessageService(messageRepo, followRepo)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)

	// ── Controllers ───────────────────────────────────────────────────────────
	return &AppContainer{
		UserRepo:      userRepo,
		NotifSvc:      notifSvc,
		MessageSvc:    messageSvc,
		AuthCtrl:      controllers.NewAuthController(authSvc),
		ProfileCtrl:   controllers.NewProfileController(profileSvc),
		DirectoryCtrl: controllers.NewDirectoryController(profileSvc, portfolioSvc),
		CompanyCtrl:   controllers.NewCompanyController(companySvc),
		PortfolioCtrl: controllers.NewPortfolioController(portfolioSvc),
		FeedCtrl:      controllers.NewFeedController(feedSvc),
		GroupCtrl:     controllers.NewGroupController(groupSvc),
		EventCtrl:     controllers.NewEventController(eventSvc),
		JobCtrl:       controllers.NewJobController(jobSvc),
		ReportCtrl:    controllers.NewReportController(reportSvc),
		AdminCtrl:     controllers.NewAdminController(adminSvc),
		MentorCtrl:    controllers.NewMentorController(mentorSvc),
		FollowCtrl:    controllers.NewFollowController(followSvc),
		MessageCtrl:   controllers.NewMessageController(messageSvc),
		NotifCtrl:     controllers.NewNotificationController(notifSvc),
		ActivityCtrl:  controllers.NewActivityController(eventSvc, jobSvc, groupSvc),
		AnalyticsCtrl: controllers.NewAnalyticsController(analyticsSvc),
	}
}
