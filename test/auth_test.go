package test

import (
	"net/http"
	"testing"
)

// TestAuth covers user registration and login.
// Run: go test -v -run TestAuth ./test/...
func TestAuth(t *testing.T) {
	sfx := newSuffix()
	email := "auth+" + sfx + "@test.com"

	t.Run("register_alumni", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/auth/register", "", map[string]any{
			"name": "Auth User", "email": email, "password": "secret123",
			"role": "alumni", "faculty": "Engineering", "major": "CS", "year_enroll": 2022,
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Alumni registered")
	})

	t.Run("register_partner", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/auth/register", "", map[string]any{
			"name": "Partner User", "email": "partner+" + sfx + "@test.com",
			"password": "secret123", "role": "partner", "company_name": "PT Test Corp",
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Partner registered")
	})

	t.Run("duplicate_email_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/auth/register", "", map[string]any{
			"name": "Dup", "email": email, "password": "secret123",
			"role": "alumni", "faculty": "X", "major": "Y", "year_enroll": 2020,
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Duplicate email rejected")
	})

	var token string

	t.Run("login_success", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/auth/login", "", map[string]any{
			"email": email, "password": "secret123",
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		token = d["token"].(string)
		if token == "" {
			t.Fatal("token not returned")
		}
		t.Log("✅ Login successful")
	})

	t.Run("login_wrong_password", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/auth/login", "", map[string]any{
			"email": email, "password": "wrongpassword",
		})
		assertStatus(t, 401, code, res)
		t.Log("✅ Wrong password rejected")
	})

	t.Run("get_me", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/me", token, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ /me returns current user")
	})

	t.Run("get_me_without_token", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/me", "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ /me without token returns 401")
	})
}
