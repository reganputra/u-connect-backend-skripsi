package main

import (
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/reganputra/skripsi-backend/models"
)

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
