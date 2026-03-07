package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestGroup covers group lifecycle, membership, articles, nested comments, and reactions.
// Run: go test -v -run TestGroup ./test/...
func TestGroup(t *testing.T) {
	sfx := newSuffix()
	ownerToken, _ := registerAndLogin(t, sfx, "Group Owner", "owner+"+sfx+"@test.com", "alumni", "Engineering", "CS")
	memberToken, memberID := registerAndLogin(t, sfx, "Group Member", "member+"+sfx+"@test.com", "alumni", "Science", "Physics")
	strangerToken, _ := registerAndLogin(t, sfx, "Stranger", "stranger+"+sfx+"@test.com", "alumni", "Law", "Law")

	var groupID, articleID, commentID, replyID float64

	// ── Group CRUD ──────────────────────────────────────────────────────────────

	t.Run("create_group", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/groups", ownerToken, map[string]string{
			"title":       "Test Group " + sfx,
			"category":    "Technology",
			"description": "Automated test group",
			"rules":       "Be respectful.",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		groupID = d["ID"].(float64)
		t.Logf("✅ Group created (id: %.0f)", groupID)
	})

	t.Run("browse_groups", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/groups", ownerToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Groups list fetched")
	})

	t.Run("get_group_detail", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/groups/%.0f", groupID), memberToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Group detail fetched (non-member can view)")
	})

	t.Run("update_group", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/groups/%.0f", groupID), ownerToken, map[string]string{
			"description": "Updated description",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Group updated (owner)")
	})

	t.Run("non_owner_cannot_update_group", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/groups/%.0f", groupID), memberToken, map[string]string{
			"title": "hacked",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner update blocked (403)")
	})

	// ── Membership ──────────────────────────────────────────────────────────────

	t.Run("member_joins_group", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/%.0f/join", groupID), memberToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Member joined group")
	})

	t.Run("cannot_join_twice", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/%.0f/join", groupID), memberToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Duplicate join rejected")
	})

	t.Run("get_members", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/groups/%.0f/members", groupID), ownerToken, nil)
		assertStatus(t, 200, code, res)
		members := res["data"].([]any)
		if len(members) != 2 {
			t.Fatalf("expected 2 members, got %d", len(members))
		}
		t.Log("✅ Members list: 2 (owner + member)")
	})

	t.Run("get_joined_groups", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/groups/joined", memberToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Joined groups list fetched")
	})

	t.Run("owner_cannot_leave", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/groups/%.0f/leave", groupID), ownerToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Owner leave blocked (400)")
	})

	// ── Articles ─────────────────────────────────────────────────────────────────

	t.Run("stranger_cannot_create_article", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, fmt.Sprintf("/api/groups/%.0f/articles", groupID), strangerToken, map[string]string{
			"title": "hacked", "content": "should fail",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-member article creation blocked (403)")
	})

	t.Run("member_creates_article", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, fmt.Sprintf("/api/groups/%.0f/articles", groupID), memberToken, map[string]string{
			"title":   "Introduction to Go " + sfx,
			"content": "Go is a statically typed compiled language...",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		articleID = d["ID"].(float64)
		t.Logf("✅ Article created by member (id: %.0f)", articleID)
	})

	t.Run("update_article", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/groups/articles/%.0f", articleID), memberToken, map[string]string{
			"content": "Updated article content",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Article updated")
	})

	// ── Comments (nested) ─────────────────────────────────────────────────────────

	t.Run("add_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/articles/%.0f/comments", articleID), ownerToken, map[string]any{
			"content": "Excellent article!",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		commentID = d["ID"].(float64)
		t.Logf("✅ Comment added (id: %.0f)", commentID)
	})

	t.Run("reply_to_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/articles/%.0f/comments", articleID), memberToken, map[string]any{
			"content":           "Thank you!",
			"parent_comment_id": commentID,
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		replyID = d["ID"].(float64)
		t.Logf("✅ Reply added (id: %.0f)", replyID)
	})

	t.Run("reply_to_reply", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/articles/%.0f/comments", articleID), ownerToken, map[string]any{
			"content":           "You're welcome!",
			"parent_comment_id": replyID,
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Reply-to-reply added (infinite nesting)")
	})

	t.Run("get_article_detail_with_nested_tree", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/groups/articles/%.0f", articleID), ownerToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		comments := d["comments"].([]any)
		if len(comments) == 0 {
			t.Fatal("expected at least one comment")
		}
		topComment := comments[0].(map[string]any)
		replies := topComment["replies"].([]any)
		if len(replies) == 0 {
			t.Fatal("expected nested reply")
		}
		t.Logf("✅ Article detail: %d comment(s) with %d reply(ies)", len(comments), len(replies))
	})

	// ── Reactions ─────────────────────────────────────────────────────────────────

	t.Run("react_to_article", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/articles/%.0f/react", articleID), ownerToken, map[string]any{
			"type": "love",
		})
		assertStatus(t, 200, code, res)
		if dataMap(res)["action"] != "added" {
			t.Fatal("expected action=added")
		}
		t.Log("✅ Article reaction added")
	})

	t.Run("stranger_cannot_react", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/articles/%.0f/react", articleID), strangerToken, map[string]any{
			"type": "like",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-member reaction blocked (403)")
	})

	t.Run("react_to_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/comments/%.0f/react", commentID), memberToken, map[string]any{
			"type": "like",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Comment reaction added")
	})

	t.Run("react_to_reply", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/comments/%.0f/react", replyID), ownerToken, map[string]any{
			"type": "haha",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Reply reaction added")
	})

	// ── Kick & Leave ──────────────────────────────────────────────────────────────

	t.Run("kick_member", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/groups/%.0f/members/%.0f", groupID, memberID), ownerToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Member kicked")
	})

	t.Run("kicked_member_cannot_create_article", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, fmt.Sprintf("/api/groups/%.0f/articles", groupID), memberToken, map[string]string{
			"title": "after kick", "content": "should fail",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Kicked member blocked (403)")
	})

	// ── Delete Group (cascade) ────────────────────────────────────────────────────

	t.Run("delete_group", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/groups/%.0f", groupID), ownerToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Group deleted")
	})

	t.Run("deleted_group_returns_404", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/groups/%.0f", groupID), ownerToken, nil)
		assertStatus(t, 404, code, res)
		t.Log("✅ Deleted group returns 404")
	})

	_ = strangerToken
}
