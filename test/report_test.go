package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestReport covers user-side content reporting.
// Run: go test -v -run TestReport ./test/...
func TestReport(t *testing.T) {
	sfx := newSuffix()

	reporterToken, _ := registerAndLogin(t, sfx, "Reporter User", "reporter+"+sfx+"@test.com", "alumni", "Engineering", "CS")
	authorToken, _ := registerAndLogin(t, sfx, "Author User", "author+"+sfx+"@test.com", "alumni", "Science", "Physics")

	var postID, commentID, reportID float64

	// ── Setup: create content to report ───────────────────────────────────────

	t.Run("setup_create_post", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/feed", authorToken, map[string]string{
			"title":   "Post to be reported " + sfx,
			"content": "This post will be reported.",
		})
		assertStatus(t, 201, code, res)
		postID = safeID(t, res, "post")
		t.Logf("✅ Post created (id: %.0f)", postID)
	})

	t.Run("setup_create_comment", func(t *testing.T) {
		if postID == 0 {
			t.Skip("⏭️  Skipped — post not created")
		}
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/comments", postID), reporterToken, map[string]any{
			"content": "A comment that will be reported",
		})
		assertStatus(t, 201, code, res)
		commentID = safeID(t, res, "comment")
		t.Logf("✅ Comment created (id: %.0f)", commentID)
	})

	// ── Report a post ─────────────────────────────────────────────────────────

	t.Run("report_post", func(t *testing.T) {
		if postID == 0 {
			t.Skip("⏭️  Skipped — post not created")
		}
		// Route: POST /api/reports (plural)
		code, res := jsonReq(t, http.MethodPost, "/api/reports", reporterToken, map[string]any{
			"target_type": "post",
			"target_id":   postID,
			"report_type": "spam",
		})
		assertStatus(t, 201, code, res)
		reportID = safeID(t, res, "report")
		t.Logf("✅ Post reported (report id: %.0f)", reportID)
	})

	t.Run("duplicate_report_rejected", func(t *testing.T) {
		if postID == 0 {
			t.Skip("⏭️  Skipped — post not created")
		}
		code, res := jsonReq(t, http.MethodPost, "/api/reports", reporterToken, map[string]any{
			"target_type": "post",
			"target_id":   postID,
			"report_type": "spam",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Duplicate pending report rejected")
	})

	// ── Report a comment ──────────────────────────────────────────────────────

	t.Run("report_comment", func(t *testing.T) {
		if commentID == 0 {
			t.Skip("⏭️  Skipped — comment not created")
		}
		code, res := jsonReq(t, http.MethodPost, "/api/reports", authorToken, map[string]any{
			"target_type": "comment",
			"target_id":   commentID,
			"report_type": "harassment",
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Comment reported")
	})

	// ── Validation cases ───────────────────────────────────────────────────────

	t.Run("invalid_target_type_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/reports", reporterToken, map[string]any{
			"target_type": "unknown_type",
			"target_id":   1,
			"report_type": "spam",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid target_type rejected")
	})

	t.Run("invalid_report_type_rejected", func(t *testing.T) {
		if postID == 0 {
			t.Skip("⏭️  Skipped — post not created")
		}
		code, res := jsonReq(t, http.MethodPost, "/api/reports", reporterToken, map[string]any{
			"target_type": "post",
			"target_id":   postID,
			"report_type": "bad_type",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid report_type rejected")
	})

	t.Run("other_report_type_requires_description", func(t *testing.T) {
		if postID == 0 {
			t.Skip("⏭️  Skipped — post not created")
		}
		code, res := jsonReq(t, http.MethodPost, "/api/reports", reporterToken, map[string]any{
			"target_type": "post",
			"target_id":   postID,
			"report_type": "other",
			// missing description
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ 'other' without description rejected")
	})

	t.Run("other_report_type_with_description_accepted", func(t *testing.T) {
		// Create a fresh post so it's not blocked by duplicate-report guard
		_, pRes := formReq(t, http.MethodPost, "/api/feed", authorToken, map[string]string{
			"title": "Another post " + sfx, "content": "Content",
		})
		otherPostID := safeID(t, pRes, "other post")

		code, res := jsonReq(t, http.MethodPost, "/api/reports", reporterToken, map[string]any{
			"target_type": "post",
			"target_id":   otherPostID,
			"report_type": "other",
			"description": "Custom reason for reporting this content",
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ 'other' with description accepted")
	})

	t.Run("unauthenticated_cannot_report", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/reports", "", map[string]any{
			"target_type": "post", "target_id": 1, "report_type": "spam",
		})
		assertStatus(t, 401, code, res)
		t.Log("✅ Unauthenticated report rejected (401)")
	})

	// ── View own reports ──────────────────────────────────────────────────────

	t.Run("view_my_reports", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/reports/mine?page=1&limit=10", reporterToken, nil)
		assertStatus(t, 200, code, res)
		if res["data"] == nil {
			t.Fatal("expected data field in response")
		}
		t.Log("✅ My reports fetched")
	})

	_ = reportID
}
