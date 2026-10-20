package config

import (
	"log"
	"os"

	"github.com/reganputra/skripsi-backend/constant"
	"github.com/reganputra/skripsi-backend/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RunMigrations performs synchronous schema migration before the HTTP server
// starts accepting traffic (C-1). It replaces the previous background
// AutoMigrate goroutine so orchestrator readiness probes are accurate.
func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(
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
		return err
	}
	log.Println("✅ Database migrated successfully!")

	if err := cleanupDuplicateFollows(db); err != nil {
		log.Printf("❌ Failed to clean duplicate follows: %v", err)
		return err
	}
	// C-1: partial unique index for active follows (soft-delete aware).
	// GORM's uniqueIndex:idx_follow_pair is a plain unique index; this
	// partial index enforces uniqueness only for non-deleted rows.
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_follows_pair_active ON follows(follower_id, following_id) WHERE deleted_at IS NULL`).Error; err != nil {
		log.Printf("❌ Failed to create ux_follows_pair_active: %v", err)
		return err
	}

	seedGeneralCategory(db)
	seedAdmin(db)

	if err := backfillMissingPartnerProfiles(db); err != nil {
		return err
	}

	return nil
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

func backfillMissingPartnerProfiles(db *gorm.DB) error {
	var missingUserIDs []uint
	if err := db.
		Table("users as u").
		Select("u.id").
		Joins("LEFT JOIN user_profiles up ON up.user_id = u.id").
		Where("u.role = ? AND up.user_id IS NULL", constant.RolePartner).
		Scan(&missingUserIDs).Error; err != nil {
		return err
	}
	for _, userID := range missingUserIDs {
		profile := &models.UserProfile{UserID: userID}
		if err := db.Where(models.UserProfile{UserID: userID}).FirstOrCreate(profile).Error; err != nil {
			return err
		}
	}
	if len(missingUserIDs) > 0 {
		log.Printf("✅ Backfilled %d missing partner profiles", len(missingUserIDs))
	}
	return nil
}
