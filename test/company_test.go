package test

import (
	"net/http"
	"testing"
)

// TestCompany covers company profile CRUD for partner-role users.
// Run: go test -v -run TestCompany ./test/...
func TestCompany(t *testing.T) {
	sfx := newSuffix()
	partnerToken, _ := registerAndLoginPartner(t,
		"Partner Corp "+sfx,
		"partner+"+sfx+"@test.com",
		"PT Maju Bersama "+sfx,
	)
	alumniToken, _ := registerAndLogin(t, sfx,
		"Alumni User", "alumni+"+sfx+"@test.com",
		"alumni", "Engineering", "CS",
	)

	// ── Role Guard ─────────────────────────────────────────────────────────────

	t.Run("alumni_cannot_access_company", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/company", alumniToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Alumni blocked from /api/company (403)")
	})

	t.Run("no_token_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/company", "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ Unauthenticated request rejected (401)")
	})

	// ── Company Profile CRUD ───────────────────────────────────────────────────

	t.Run("create_company_profile", func(t *testing.T) {
		size := 50
		code, res := jsonReq(t, http.MethodPost, "/api/company", partnerToken, map[string]any{
			"industry_type": "Technology",
			"location":      "Jakarta",
			"employee_size": size,
			"website_url":   "https://majubersama.co.id",
		})
		// 201 when newly created, 200 when joining an existing company profile
		if code != http.StatusCreated && code != http.StatusOK {
			assertStatus(t, 201, code, res)
		}
		t.Logf("✅ Company profile created/joined (status %d)", code)
	})

	t.Run("get_company_profile", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/company", partnerToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		// Models have no json tags → serialized as PascalCase by encoding/json
		if d["CompanyName"] == nil {
			t.Fatal("CompanyName missing from profile response")
		}
		t.Logf("✅ Company profile fetched: %v", d["CompanyName"])
	})

	t.Run("update_company_profile", func(t *testing.T) {
		size := 100
		code, res := jsonReq(t, http.MethodPut, "/api/company", partnerToken, map[string]any{
			"location":      "Bandung",
			"employee_size": size,
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		// Models have no json tags → serialized as PascalCase by encoding/json
		if d["Location"] != "Bandung" {
			t.Fatalf("expected Location=Bandung, got %v", d["Location"])
		}
		t.Log("✅ Company profile updated")
	})

	t.Run("negative_employee_size_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, "/api/company", partnerToken, map[string]any{
			"employee_size": -5,
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Negative employee_size rejected (400)")
	})

	t.Run("second_partner_joins_same_company", func(t *testing.T) {
		// Register a second partner with the same company name → should join (200)
		sfx2 := newSuffix()
		token2, _ := registerAndLoginPartner(t,
			"Partner Corp 2 "+sfx2,
			"partner2+"+sfx2+"@test.com",
			"PT Maju Bersama "+sfx, // same company name as first partner
		)
		code, res := jsonReq(t, http.MethodPost, "/api/company", token2, map[string]any{})
		assertStatus(t, 200, code, res)
		t.Log("✅ Second partner with same company name joins existing profile (200)")
	})

	// ── Cleanup ────────────────────────────────────────────────────────────────

	t.Run("delete_company_profile", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, "/api/company", partnerToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Company profile deleted")
	})

	t.Run("get_deleted_profile_returns_404", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/company", partnerToken, nil)
		assertStatus(t, 404, code, res)
		t.Log("✅ Deleted company profile returns 404")
	})
}
