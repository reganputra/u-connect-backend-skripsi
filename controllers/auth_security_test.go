package controllers

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
)

type cookieAuthService struct {
	service.AuthService
	seen string
}

func (s *cookieAuthService) Login(service.LoginRequest) (*service.LoginResponse, error) {
	return &service.LoginResponse{Token: "access", RefreshToken: "private-refresh"}, nil
}
func (s *cookieAuthService) Refresh(raw string) (*service.LoginResponse, error) {
	s.seen = raw
	return &service.LoginResponse{Token: "access", RefreshToken: "private-refresh"}, nil
}

func TestRefreshCookieOnly(t *testing.T) {
	for _, tc := range []struct{ name, body, header, cookie, want string }{
		{"cookie", "", "", "private-refresh", "private-refresh"},
		{"body ignored", `{"refresh_token":"body-secret"}`, "", "", ""},
		{"header ignored", "", "header-secret", "", ""},
		{"cookie takes precedence", `{"refresh_token":"body-secret"}`, "header-secret", "private-refresh", "private-refresh"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &cookieAuthService{}
			ctrl := NewAuthController(svc)
			app := fiber.New()
			app.Post("/", ctrl.Refresh)
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Refresh-Token", tc.header)
			if tc.cookie != "" {
				req.Header.Set("Cookie", "refresh_token="+tc.cookie)
			}
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if svc.seen != tc.want {
				t.Fatalf("service received %q want %q", svc.seen, tc.want)
			}
		})
	}
}

func TestLoginRefreshCookieSecurity(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	for _, route := range []string{"login", "refresh"} {
		t.Run(route, func(t *testing.T) {
			svc := &cookieAuthService{}
			ctrl := NewAuthController(svc)
			app := fiber.New()
			app.Post("/login", ctrl.Login)
			app.Post("/refresh", ctrl.Refresh)
			req := httptest.NewRequest("POST", "/"+route, strings.NewReader(`{"email":"a","password":"b"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Cookie", "refresh_token=old")
			res, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				t.Fatalf("status %d", res.StatusCode)
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(body), "private-refresh") || strings.Contains(string(body), "refresh_token") {
				t.Fatal("refresh leaked in response")
			}
			cookies := res.Cookies()
			if len(cookies) != 1 {
				t.Fatalf("cookies=%d", len(cookies))
			}
			c := cookies[0]
			if c.Name != "refresh_token" || !c.HttpOnly || !c.Secure || c.MaxAge != 604800 || c.Path != "/" || c.SameSite != 3 {
				t.Fatalf("insecure refresh cookie: %+v", c)
			}
			if res.Header.Get("Cache-Control") != "no-store" {
				t.Fatal("token response may be cached")
			}
		})
	}
}
