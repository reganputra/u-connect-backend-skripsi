package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestProfile covers profile CRUD and experience management.
// Run: go test -v -run TestProfile ./test/...
func TestProfile(t *testing.T) {
	sfx := newSuffix()
	token, _ := registerAndLogin(t, sfx,
		"Profile Tester", "profile+"+sfx+"@test.com",
		"alumni", "Engineering", "Informatics",
	)

	t.Run("create_profile_employed", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/profile", token, map[string]string{
			"bio":          "Software engineer with 3 years experience",
			"location":     "Jakarta",
			"job_status":   "employed",
			"position":     "Backend Engineer",
			"company_name": "PT Tech Indonesia",
			"salary":       "15000000",
			"skills":       "Go, PostgreSQL, Docker",
			"interests":    "Backend, Cloud",
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Profile created (employed)")
	})

	t.Run("get_profile", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/profile", token, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		// Models have no json tags → serialized as PascalCase by encoding/json
		if d["JobStatus"] == nil {
			t.Fatal("JobStatus missing from profile response")
		}
		t.Log("✅ Profile fetched")
	})

	t.Run("update_profile_bio", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, "/api/profile", token, map[string]string{
			"bio": "Updated bio from test suite",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Profile bio updated")
	})

	t.Run("update_job_status_entrepreneur", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, "/api/profile", token, map[string]string{
			"job_status":    "entrepreneur",
			"industry_name": "AgriTech",
			"industry_type": "B2B",
			"year_founding": "2023",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Profile updated to entrepreneur")
	})

	var expID float64

	t.Run("add_experience", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/profile/experience", token, map[string]any{
			"company_name": "PT Previous Job",
			"position":     "Junior Developer",
			"start_year":   2020,
			"end_year":     2022,
			"description":  "Worked on REST APIs",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		expID = d["ID"].(float64)
		t.Logf("✅ Experience added (id: %.0f)", expID)
	})

	t.Run("update_experience", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, fmt.Sprintf("/api/profile/experience/%.0f", expID), token, map[string]any{
			"description": "Updated description",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Experience updated")
	})

	t.Run("delete_experience", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/profile/experience/%.0f", expID), token, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Experience deleted")
	})

	t.Run("delete_profile", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, "/api/profile", token, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Profile deleted")
	})

	t.Run("get_deleted_profile_returns_404", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/profile", token, nil)
		assertStatus(t, 404, code, res)
		t.Log("✅ Deleted profile returns 404")
	})
}
