package ws

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
)

type wsUsers struct {
	repository.UserRepository
	user *models.User
}

func (r wsUsers) FindUserByID(id uint) (*models.User, error) { return r.user, nil }

func TestWSAuth(t *testing.T) {
	t.Setenv("JWT_SECRET", "unit-test-secret-only")
	for _, tc := range []struct {
		name, transport, purpose string
		active                   bool
		version                  int
		want                     int
	}{
		{"header", "header", "access", true, 1, 200},
		{"browser protocol", "protocol", "access", true, 1, 200},
		{"query forbidden", "query", "access", true, 1, 401},
		{"disabled", "header", "access", false, 1, 401},
		{"revoked", "header", "access", true, 0, 401},
		{"refresh rejected", "header", "refresh", true, 1, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"user_id": 1, "role": "student", "token_version": tc.version, "token_type": tc.purpose, "jti": "test-jti",
				"exp": time.Now().Add(time.Minute).Unix(), "iat": time.Now().Unix(),
			}).SignedString([]byte("unit-test-secret-only"))
			if err != nil {
				t.Fatal(err)
			}
			app := fiber.New()
			repo := wsUsers{user: &models.User{Model: gorm.Model{ID: 1}, IsActive: tc.active, TokenVersion: 1, Role: "student"}}
			app.Get("/ws", WSAuthMiddleware(repo, nil), func(c *fiber.Ctx) error { return c.SendStatus(200) })
			url := "/ws"
			if tc.transport == "query" {
				url += "?token=" + token
			}
			req := httptest.NewRequest("GET", url, nil)
			if tc.transport == "header" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			if tc.transport == "protocol" {
				req.Header.Set("Sec-WebSocket-Protocol", "Bearer, "+token)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.want {
				t.Fatalf("got %d want %d", resp.StatusCode, tc.want)
			}
		})
	}
}
