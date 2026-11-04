package repository

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Opt in to a disposable PostgreSQL database, never the application's database.
func authTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("AUTH_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("AUTH_TEST_DATABASE_DSN not set (requires disposable PostgreSQL)")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	schema := "auth_test_" + uuid.NewString()[:8]
	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP SCHEMA " + schema + " CASCADE"); sqlDB, _ := db.DB(); sqlDB.Close() })
	// Each connection uses a schema-qualified table, including all repository calls.
	db = db.Scopes(func(tx *gorm.DB) *gorm.DB { return tx.Table(schema + ".users") })
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	return db.Session(&gorm.Session{})
}

func TestDBAuthAtomicity(t *testing.T) {
	db := authTestDB(t)
	repo := NewUserRepository(db)
	u := models.User{Name: "Test", Email: "auth@example.invalid", Password: "old", Role: "student", IsActive: true}
	if err := repo.CreateUser(&u); err != nil {
		t.Fatal(err)
	}
	var wins atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if repo.CompareAndSwapTokenVersion(u.ID, 0) == nil {
				wins.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("CAS winners=%d want 1", wins.Load())
	}
	if err := repo.CompareAndSwapTokenVersion(u.ID, 0); err == nil {
		t.Fatal("replay accepted")
	}
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"logout", func() error { return repo.BumpTokenVersion(u.ID) }},
		{"password", func() error { return repo.UpdateUserPassword(u.ID, "new") }},
		{"reset", func() error { return repo.ResetPasswordAndClearAttempts(u.ID, "reset") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, err := repo.FindUserByID(u.ID)
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.run(); err != nil {
				t.Fatal(err)
			}
			after, err := repo.FindUserByID(u.ID)
			if err != nil {
				t.Fatal(err)
			}
			if after.TokenVersion != before.TokenVersion+1 {
				t.Fatalf("version %d -> %d", before.TokenVersion, after.TokenVersion)
			}
		})
	}
	admin := NewAdminRepository(db)
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"disable", func() error { return admin.UpdateUserStatus(u.ID, false) }},
		{"role", func() error { return admin.UpdateUserRole(u.ID, "admin") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before, err := repo.FindUserByID(u.ID)
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.run(); err != nil {
				t.Fatal(err)
			}
			after, err := repo.FindUserByID(u.ID)
			if err != nil {
				t.Fatal(err)
			}
			if after.TokenVersion != before.TokenVersion+1 {
				t.Fatalf("version %d -> %d", before.TokenVersion, after.TokenVersion)
			}
			if after.Password != "reset" {
				t.Fatal("admin update overwrote password")
			}
		})
	}
	t.Run("concurrent reset lock", func(t *testing.T) {
		var attempts sync.WaitGroup
		for i := 0; i < 16; i++ {
			attempts.Add(1)
			go func() {
				defer attempts.Done()
				if err := repo.IncrementResetAttempts(u.ID); err != nil {
					t.Error(err)
				}
			}()
		}
		attempts.Wait()
		locked, err := repo.FindUserByID(u.ID)
		if err != nil {
			t.Fatal(err)
		}
		if locked.ResetAttempts != 16 || locked.ResetLockedAt == nil {
			t.Fatalf("reset attempts=%d locked_at=%v", locked.ResetAttempts, locked.ResetLockedAt)
		}
	})
	current, err := repo.FindUserByID(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if repo.CompareAndSwapTokenVersion(u.ID, current.TokenVersion) == nil {
		t.Fatal("disabled user rotated")
	}
}
