package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestPortfolio covers portfolio item CRUD for alumni/student-role users.
// Run: go test -v -run TestPortfolio ./test/...
func TestPortfolio(t *testing.T) {
	sfx := newSuffix()
	tokenA, _ := registerAndLogin(t, sfx,
		"Portfolio User A", "portfolioa+"+sfx+"@test.com",
		"alumni", "Engineering", "CS",
	)
	tokenB, _ := registerAndLogin(t, sfx,
		"Portfolio User B", "portfoliob+"+sfx+"@test.com",
		"alumni", "Science", "Physics",
	)
	// Partner should be blocked by the RequireRole("alumni","student") middleware
	partnerToken, _ := registerAndLoginPartner(t,
		"Portfolio Partner "+sfx,
		"portfoliopartner+"+sfx+"@test.com",
		"PT Portfolio Ltd "+sfx,
	)

	var itemID float64

	// ── Role Guard ─────────────────────────────────────────────────────────────

	t.Run("partner_cannot_access_portfolio", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/portfolio", partnerToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Partner blocked from /api/portfolio (403)")
	})

	t.Run("no_token_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/portfolio", "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ Unauthenticated request rejected (401)")
	})

	// ── Portfolio CRUD ─────────────────────────────────────────────────────────

	t.Run("create_portfolio_item", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/portfolio", tokenA, map[string]string{
			"title":       "Final Year Project " + sfx,
			"description": "Built a REST API backend with Go and Fiber",
			"category":    "Software Engineering",
			"tags":        "Go,Fiber,PostgreSQL",
			"start_date":  "2024-01",
			"end_date":    "2024-06",
			"link":        "https://github.com/example/fyp-" + sfx,
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		itemID = d["ID"].(float64)
		if d["Link"] == nil {
			t.Fatal("expected Link to be set")
		}
		t.Logf("✅ Portfolio item created (id: %.0f)", itemID)
	})

	t.Run("create_item_without_title_fails", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/portfolio", tokenA, map[string]string{
			"description": "Missing title",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Portfolio item without title rejected (400)")
	})

	t.Run("get_portfolio_items", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/portfolio", tokenA, nil)
		assertStatus(t, 200, code, res)
		items, ok := res["data"].([]any)
		if !ok || len(items) == 0 {
			t.Fatal("expected at least one portfolio item in the list")
		}
		t.Logf("✅ Portfolio items fetched (%d item(s))", len(items))
	})

	t.Run("update_portfolio_item", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/portfolio/%.0f", itemID), tokenA, map[string]string{
			"description": "Updated: deployed to production with Docker",
			"tags":        "Go,Fiber,PostgreSQL,Docker",
			"link":        "https://demo.example.com/" + sfx,
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Link"] == nil {
			t.Fatal("expected Link to remain set after update")
		}
		t.Log("✅ Portfolio item updated")
	})

	t.Run("cannot_update_another_users_item", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/portfolio/%.0f", itemID), tokenB, map[string]string{
			"title": "hijacked",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Cross-user update blocked (403)")
	})

	t.Run("cannot_delete_another_users_item", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/portfolio/%.0f", itemID), tokenB, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Cross-user delete blocked (403)")
	})

	t.Run("other_user_portfolio_is_empty", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/portfolio", tokenB, nil)
		assertStatus(t, 200, code, res)
		// tokenB has no items — data should be an empty slice or null
		if items, ok := res["data"].([]any); ok && len(items) != 0 {
			t.Fatalf("expected 0 items for user B, got %d", len(items))
		}
		t.Log("✅ User B has empty portfolio")
	})

	// ── Cleanup ────────────────────────────────────────────────────────────────

	t.Run("delete_portfolio_item", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/portfolio/%.0f", itemID), tokenA, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Portfolio item deleted")
	})

	t.Run("get_after_delete_returns_empty_list", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/portfolio", tokenA, nil)
		assertStatus(t, 200, code, res)
		// After deletion the list should be empty
		if items, ok := res["data"].([]any); ok && len(items) != 0 {
			t.Fatalf("expected 0 items after deletion, got %d", len(items))
		}
		t.Log("✅ Portfolio list empty after deletion")
	})
}
