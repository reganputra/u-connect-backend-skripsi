package middleware

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"net/http/httptest"
	"testing"
	"time"
)

type guardRepo struct {
	repository.UserRepository
	user *models.User
	err  error
}

func (r guardRepo) FindUserByID(uint) (*models.User, error) { return r.user, r.err }

func TestProtectedSecurity(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-only-key")
	for _, tc := range []struct {
		name    string
		repo    repository.UserRepository
		version interface{}
		purpose string
		want    int
	}{
		{"valid", guardRepo{user: &models.User{IsActive: true, Role: "student", TokenVersion: 2}}, 2, "access", 200},
		{"missing dependency", nil, 2, "access", 401},
		{"disabled", guardRepo{user: &models.User{IsActive: false, Role: "student", TokenVersion: 2}}, 2, "access", 401},
		{"revoked", guardRepo{user: &models.User{IsActive: true, Role: "student", TokenVersion: 3}}, 2, "access", 401},
		{"role changed", guardRepo{user: &models.User{IsActive: true, Role: "admin", TokenVersion: 2}}, 2, "access", 401},
		{"database error", guardRepo{err: errors.New("offline")}, 2, "access", 401},
		{"refresh", guardRepo{user: &models.User{IsActive: true, Role: "student", TokenVersion: 2}}, 2, "refresh", 401},
		{"missing version", guardRepo{user: &models.User{IsActive: true, Role: "student"}}, nil, "access", 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				if tc.repo != nil {
					c.Locals("auth_user_repository", tc.repo)
				}
				return c.Next()
			})
			app.Get("/", Protected(), func(c *fiber.Ctx) error { return c.SendStatus(200) })
			claims := jwt.MapClaims{"user_id": 1, "role": "student", "token_version": tc.version, "token_type": tc.purpose, "exp": time.Now().Add(time.Minute).Unix()}
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-only-key"))
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != tc.want {
				t.Fatalf("got %d want %d", res.StatusCode, tc.want)
			}
		})
	}
}
