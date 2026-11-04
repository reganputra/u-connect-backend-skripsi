package service

import (
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/reganputra/skripsi-backend/middleware"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDBRefreshAndRevocation(t *testing.T) {
	dsn := os.Getenv("AUTH_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("AUTH_TEST_DATABASE_DSN not set (disposable PostgreSQL required)")
	}
	t.Setenv("JWT_SECRET", "test-only-key")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	schema := "auth_svc_" + uuid.NewString()[:8]
	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP SCHEMA " + schema + " CASCADE"); sqlDB, _ := db.DB(); sqlDB.Close() })
	scoped := db.Scopes(func(tx *gorm.DB) *gorm.DB { return tx.Table(schema + ".users") })
	if err := scoped.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	scoped = scoped.Session(&gorm.Session{})
	repo := repository.NewUserRepository(scoped)
	admin := repository.NewAdminRepository(scoped)
	u := models.User{Name: "Test", Email: "auth@example.invalid", Role: "student", Password: "old", IsActive: true}
	if err := repo.CreateUser(&u); err != nil {
		t.Fatal(err)
	}
	svc := &authService{userRepo: repo}
	access, refresh, err := svc.issueTokens(&u)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Use(middleware.AuthRepository(repo))
	app.Get("/", middleware.Protected(), func(c *fiber.Ctx) error { return c.SendStatus(200) })
	check := func(token string, want int) {
		t.Helper()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != want {
			t.Fatalf("status=%d want %d", res.StatusCode, want)
		}
	}
	check(access, 200)
	check(refresh, 401)
	var wins atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan *LoginResponse, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if res, err := svc.Refresh(refresh); err == nil {
				wins.Add(1)
				results <- res
			}
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	if wins.Load() != 1 {
		t.Fatalf("refresh winners=%d want 1", wins.Load())
	}
	check(access, 401)
	winner := <-results
	check(winner.Token, 200)
	if _, err := svc.Refresh(refresh); err == nil {
		t.Fatal("refresh replay accepted")
	}
	for _, tc := range []struct {
		name   string
		mutate func() error
	}{
		{"logout", func() error { return svc.Logout(u.ID) }},
		{"password", func() error { return repo.UpdateUserPassword(u.ID, "changed") }},
		{"reset", func() error { return repo.ResetPasswordAndClearAttempts(u.ID, "reset") }},
		{"role", func() error { return admin.UpdateUserRole(u.ID, "admin") }},
		{"disable", func() error { return admin.UpdateUserStatus(u.ID, false) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current, err := repo.FindUserByID(u.ID)
			if err != nil {
				t.Fatal(err)
			}
			access, refresh, err := svc.issueTokens(current)
			if err != nil {
				t.Fatal(err)
			}
			check(access, 200)
			if err := tc.mutate(); err != nil {
				t.Fatal(err)
			}
			check(access, 401)
			if _, err := svc.Refresh(refresh); err == nil {
				t.Fatal("revoked refresh accepted")
			}
		})
	}
	// Expired lock and attempt state remain isolated from unrelated security fields.
	if err := repo.LockResetWithTime(u.ID, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
}
