package config

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/reganputra/skripsi-backend/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := strings.TrimSpace(utils.GetEnv("DATABASE_URL", ""))
	if dsn == "" {
		host := utils.GetEnv("DB_HOST", "localhost")
		port := utils.GetEnv("DB_PORT", "5432")
		user := utils.GetEnv("DB_USER", "postgres")
		password := utils.GetEnv("DB_PASSWORD", "rootpassword")
		dbname := utils.GetEnv("DB_NAME", "alumni_community_db")
		sslmode := utils.GetEnv("DB_SSLMODE", "disable")

		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
			host, port, user, password, dbname, sslmode,
		)
	} else {
		parsed, err := url.Parse(dsn)
		if err == nil {
			query := parsed.Query()
			if query.Get("sslmode") == "" {
				query.Set("sslmode", "require")
			}
			if query.Get("pgbouncer") == "" {
				query.Set("pgbouncer", "true")
			}
			parsed.RawQuery = query.Encode()
			dsn = parsed.String()
		}
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get underlying sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}

	// ── Connection Pool Configuration ─────────────────────────────────────────
	// Limit concurrent DB connections; tune against Postgres max_connections.
	sqlDB.SetMaxOpenConns(25)                 // max simultaneous open connections
	sqlDB.SetMaxIdleConns(5)                  // keep warm connections available for bursts
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // recycle connections to avoid stale TCP
	sqlDB.SetConnMaxIdleTime(1 * time.Hour)   // reap idle connections after inactivity

	DB = db
	log.Println("✅ Successfully connected to the database!")
}
