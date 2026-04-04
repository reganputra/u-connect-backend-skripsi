package test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/reganputra/skripsi-backend/utils"
)

// loginAdmin logs in using ADMIN_EMAIL / ADMIN_PASSWORD from the .env file.
func loginAdmin(t *testing.T) string {
	t.Helper()
	utils.LoadEnvFile("../.env")

	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" {
		email = "admin@mail.com"
	}
	if password == "" {
		password = "admin123"
	}

	_, res := jsonReq(t, http.MethodPost, "/api/auth/login", "", map[string]any{
		"email": email, "password": password,
	})
	d := dataMap(res)
	token, ok := d["token"].(string)
	if !ok || token == "" {
		t.Fatal("❌ Could not login as admin — is the server running and admin seeded?")
	}
	return token
}

// TestAdmin covers dashboard, user management, report moderation, content deletion, and categories.
// Requires the server running and an admin account seeded.
// Run: go test -v -run TestAdmin ./test/...
func TestAdmin(t *testing.T) {
	adminToken := loginAdmin(t)
	sfx := newSuffix()

	// Regular alumni user for creating content to moderate
	userToken, userID := registerAndLogin(t, sfx, "Target User", "target+"+sfx+"@test.com", "alumni", "Engineering", "CS")

	var postID, groupID, eventID, jobID float64
	var reportID, categoryID float64

	// ── Setup: create content across all modules ──────────────────────────────

	t.Run("setup_post", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/feed", userToken, map[string]string{
			"title": "Admin Test Post " + sfx, "content": "content",
		})
		assertStatus(t, 201, code, res)
		postID = safeID(t, res, "post")
		t.Logf("✅ Post created (id: %.0f)", postID)
	})

	t.Run("setup_group", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/groups", userToken, map[string]string{
			"title": "Admin Test Group " + sfx, "category": "Tech",
		})
		assertStatus(t, 201, code, res)
		groupID = safeID(t, res, "group")
		t.Logf("✅ Group created (id: %.0f)", groupID)
	})

	t.Run("setup_event", func(t *testing.T) {
		// Event controller reads form values, not JSON body
		code, res := formReq(t, http.MethodPost, "/api/events", userToken, map[string]string{
			"title":    "Admin Test Event " + sfx,
			"location": "Online",
			"capacity": "50",
			"status":   "upcoming",
		})
		assertStatus(t, 201, code, res)
		eventID = safeID(t, res, "event")
		t.Logf("✅ Event created (id: %.0f)", eventID)
	})

	t.Run("setup_job", func(t *testing.T) {
		// Job controller reads form values, not JSON body
		code, res := formReq(t, http.MethodPost, "/api/jobs", userToken, map[string]string{
			"title":        "Admin Test Job " + sfx,
			"description":  "job description",
			"company_name": "PT Test Corp",
			"location":     "Jakarta",
			"job_type":     "full-time",
			"status":       "open",
		})
		assertStatus(t, 201, code, res)
		jobID = safeID(t, res, "job")
		t.Logf("✅ Job created (id: %.0f)", jobID)
	})

	t.Run("setup_report", func(t *testing.T) {
		// Create a dedicated post to be reported
		_, pRes := formReq(t, http.MethodPost, "/api/feed", userToken, map[string]string{
			"title": "Post to report " + sfx, "content": "reportable content",
		})
		reportablePostID := safeID(t, pRes, "reportable post")

		// Route is /api/reports (plural)
		code, res := jsonReq(t, http.MethodPost, "/api/reports", userToken, map[string]any{
			"target_type": "post",
			"target_id":   reportablePostID,
			"report_type": "spam",
		})
		assertStatus(t, 201, code, res)
		reportID = safeID(t, res, "report")
		t.Logf("✅ Report created (id: %.0f)", reportID)
	})

	// ── Dashboard ─────────────────────────────────────────────────────────────

	t.Run("get_dashboard", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/admin/dashboard", adminToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		for _, key := range []string{"users", "posts", "groups", "events", "jobs", "reports_pending"} {
			if d[key] == nil {
				t.Fatalf("❌ missing stat key: %s", key)
			}
		}
		t.Logf("✅ Dashboard: users=%.0f posts=%.0f groups=%.0f events=%.0f jobs=%.0f pending=%.0f",
			d["users"], d["posts"], d["groups"], d["events"], d["jobs"], d["reports_pending"])
	})

	t.Run("non_admin_cannot_access_dashboard", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/admin/dashboard", userToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-admin blocked from dashboard (403)")
	})

	t.Run("unauthenticated_blocked", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/admin/dashboard", "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ No token blocked (401)")
	})

	// ── User Management ───────────────────────────────────────────────────────

	t.Run("list_all_users", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/admin/users?page=1&limit=10", adminToken, nil)
		assertStatus(t, 200, code, res)
		if res["data"] == nil {
			t.Fatal("expected data field")
		}
		t.Log("✅ All users listed")
	})

	t.Run("list_users_filtered_by_role", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/admin/users?role=alumni", adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Users filtered by role=alumni")
	})

	t.Run("get_user_by_id", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/admin/users/%.0f", userID), adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ User detail fetched")
	})

	t.Run("deactivate_user", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/users/%.0f/status", userID), adminToken, map[string]any{
			"is_active": false,
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["IsActive"] != false {
			t.Fatalf("expected IsActive=false, got %v", d["IsActive"])
		}
		t.Log("✅ User deactivated")
	})

	t.Run("reactivate_user", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/users/%.0f/status", userID), adminToken, map[string]any{
			"is_active": true,
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["IsActive"] != true {
			t.Fatalf("expected IsActive=true, got %v", d["IsActive"])
		}
		t.Log("✅ User reactivated")
	})

	t.Run("change_user_role", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/users/%.0f/role", userID), adminToken, map[string]any{
			"role": "student",
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Role"] != "student" {
			t.Fatalf("expected Role=student, got %v", d["Role"])
		}
		t.Log("✅ User role changed to student")
	})

	t.Run("invalid_role_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/users/%.0f/role", userID), adminToken, map[string]any{
			"role": "superuser",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid role rejected")
	})

	t.Run("restore_user_role", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/users/%.0f/role", userID), adminToken, map[string]any{
			"role": "alumni",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ User role restored to alumni")
	})

	// ── Report Moderation ─────────────────────────────────────────────────────

	t.Run("list_all_reports", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/admin/reports?page=1&limit=10", adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ All reports listed")
	})

	t.Run("list_pending_reports", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/admin/reports?status=pending", adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Pending reports listed")
	})

	t.Run("get_report_by_id", func(t *testing.T) {
		if reportID == 0 {
			t.Skip("⏭️  Skipped — setup_report failed (reportID not set)")
		}
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/admin/reports/%.0f", reportID), adminToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Status"] == nil {
			t.Fatal("expected Status in report")
		}
		t.Logf("✅ Report detail fetched (status: %v)", d["Status"])
	})

	t.Run("reject_report", func(t *testing.T) {
		if reportID == 0 {
			t.Skip("⏭️  Skipped — setup_report failed (reportID not set)")
		}
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/reports/%.0f/reject", reportID), adminToken, map[string]any{
			"admin_note": "Report does not violate our policies.",
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Status"] != "rejected" {
			t.Fatalf("expected status=rejected, got %v", d["Status"])
		}
		t.Log("✅ Report rejected with reason")
	})

	t.Run("cannot_reject_already_processed_report", func(t *testing.T) {
		if reportID == 0 {
			t.Skip("⏭️  Skipped — setup_report failed")
		}
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/reports/%.0f/reject", reportID), adminToken, map[string]any{
			"admin_note": "try again",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Re-processing already-resolved report blocked")
	})

	t.Run("resolve_report_with_content_deletion", func(t *testing.T) {
		// Create a fresh post + report — independent of earlier setup
		_, pRes := formReq(t, http.MethodPost, "/api/feed", userToken, map[string]string{
			"title": "Delete me " + sfx, "content": "bad content",
		})
		badPostID := safeID(t, pRes, "delete-me post")

		_, rRes := jsonReq(t, http.MethodPost, "/api/reports", userToken, map[string]any{ // /api/reports (plural)
			"target_type": "post", "target_id": badPostID, "report_type": "inappropriate",
		})
		newReportID := safeID(t, rRes, "new report")

		note := "Content deleted — violated community guidelines"
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/reports/%.0f/resolve", newReportID), adminToken, map[string]any{
			"admin_note":     note,
			"delete_content": true,
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Status"] != "resolved" {
			t.Fatalf("expected status=resolved, got %v", d["Status"])
		}
		t.Logf("✅ Report resolved + post %.0f deleted", badPostID)
	})

	// ── Direct Content Deletion ────────────────────────────────────────────────

	t.Run("admin_delete_post", func(t *testing.T) {
		if postID == 0 {
			t.Skip("⏭️  Skipped — setup_post failed")
		}
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/admin/posts/%.0f", postID), adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Logf("✅ Admin deleted post %.0f", postID)
	})

	t.Run("admin_delete_group", func(t *testing.T) {
		if groupID == 0 {
			t.Skip("⏭️  Skipped — setup_group failed")
		}
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/admin/groups/%.0f", groupID), adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Logf("✅ Admin deleted group %.0f", groupID)
	})

	t.Run("admin_delete_event", func(t *testing.T) {
		if eventID == 0 {
			t.Skip("⏭️  Skipped — setup_event failed")
		}
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/admin/events/%.0f", eventID), adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Logf("✅ Admin deleted event %.0f", eventID)
	})

	t.Run("admin_delete_job", func(t *testing.T) {
		if jobID == 0 {
			t.Skip("⏭️  Skipped — setup_job failed")
		}
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/admin/jobs/%.0f", jobID), adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Logf("✅ Admin deleted job %.0f", jobID)
	})

	// ── Category Management ────────────────────────────────────────────────────

	t.Run("create_category", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/admin/categories", adminToken, map[string]any{
			"name":        "Test Category " + sfx,
			"description": "Created by admin test suite",
		})
		assertStatus(t, 201, code, res)
		categoryID = safeID(t, res, "category")
		t.Logf("✅ Category created (id: %.0f)", categoryID)
	})

	t.Run("get_categories_public", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/categories", userToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Categories public list fetched")
	})

	t.Run("update_category", func(t *testing.T) {
		if categoryID == 0 {
			t.Skip("⏭️  Skipped — create_category failed")
		}
		code, res := jsonReq(t, http.MethodPut, fmt.Sprintf("/api/admin/categories/%.0f", categoryID), adminToken, map[string]any{
			"name": "Updated Category " + sfx,
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Category updated")
	})

	t.Run("non_admin_cannot_create_category", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/admin/categories", userToken, map[string]any{
			"name": "hacked category",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-admin category creation blocked (403)")
	})

	t.Run("duplicate_category_name_rejected", func(t *testing.T) {
		if categoryID == 0 {
			t.Skip("⏭️  Skipped — create_category failed")
		}
		code, res := jsonReq(t, http.MethodPost, "/api/admin/categories", adminToken, map[string]any{
			"name": "Updated Category " + sfx, // same name as updated above
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Duplicate category name rejected")
	})

	t.Run("delete_category", func(t *testing.T) {
		if categoryID == 0 {
			t.Skip("⏭️  Skipped — create_category failed")
		}
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/admin/categories/%.0f", categoryID), adminToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Category deleted")
	})

	t.Run("delete_nonexistent_category", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, "/api/admin/categories/99999", adminToken, nil)
		assertStatus(t, 404, code, res)
		t.Log("✅ Deleting non-existent category returns 404")
	})

	_ = userID
}
