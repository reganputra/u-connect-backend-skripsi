package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8080"

// newSuffix returns a unique string per test run to avoid duplicate emails.
func newSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

// ─── HTTP Helpers ─────────────────────────────────────────────────────────────

// jsonReq sends a JSON request and returns (statusCode, parsed body).
func jsonReq(t *testing.T, method, path, token string, body any) (int, map[string]any) {
	t.Helper()
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req, _ := http.NewRequest(method, baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return resp.StatusCode, result
}

// formReq sends a multipart/form-data request and returns (statusCode, parsed body).
func formReq(t *testing.T, method, path, token string, fields map[string]string) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	w.Close()
	req, _ := http.NewRequest(method, baseURL+path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("❌ HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return resp.StatusCode, result
}

// dataMap extracts the "data" field as map from a response.
func dataMap(result map[string]any) map[string]any {
	if d, ok := result["data"].(map[string]any); ok {
		return d
	}
	return map[string]any{}
}

// assertStatus fails the test if the status code doesn't match, printing the full response.
func assertStatus(t *testing.T, want, got int, result map[string]any) {
	t.Helper()
	if want != got {
		b, _ := json.MarshalIndent(result, "", "  ")
		t.Fatalf("expected status %d, got %d\nresponse:\n%s", want, got, b)
	}
}

// registerAndLogin is a convenience helper that registers a user and returns their token and user ID.
func registerAndLogin(t *testing.T, suffix, name, email, role, faculty, major string) (token string, userID float64) {
	t.Helper()
	// Auth endpoints use JSON body (BodyParser with json tags only)
	_, regRes := jsonReq(t, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name": name, "email": email, "password": "secret123",
		"role": role, "faculty": faculty, "major": major, "year_enroll": 2022,
	})
	regData := dataMap(regRes)
	if id, ok := regData["id"].(float64); ok {
		userID = id
	}

	_, loginRes := jsonReq(t, http.MethodPost, "/api/auth/login", "", map[string]any{
		"email": email, "password": "secret123",
	})
	loginData := dataMap(loginRes)
	token = loginData["token"].(string)
	return
}

// registerAndLoginPartner registers a partner user and returns their token and user ID.
func registerAndLoginPartner(t *testing.T, name, email, companyName string) (token string, userID float64) {
	t.Helper()
	_, regRes := jsonReq(t, http.MethodPost, "/api/auth/register", "", map[string]any{
		"name": name, "email": email, "password": "secret123",
		"role": "partner", "company_name": companyName,
	})
	regData := dataMap(regRes)
	if id, ok := regData["id"].(float64); ok {
		userID = id
	}

	_, loginRes := jsonReq(t, http.MethodPost, "/api/auth/login", "", map[string]any{
		"email": email, "password": "secret123",
	})
	loginData := dataMap(loginRes)
	token = loginData["token"].(string)
	return
}
