package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestFeed covers posts, comments, nested replies, reactions, and votes.
// Run: go test -v -run TestFeed ./test/...
func TestFeed(t *testing.T) {
	sfx := newSuffix()
	tokenA, _ := registerAndLogin(t, sfx, "Feed User A", "feeda+"+sfx+"@test.com", "alumni", "Engineering", "CS")
	tokenB, _ := registerAndLogin(t, sfx, "Feed User B", "feedb+"+sfx+"@test.com", "alumni", "Science", "Physics")

	var postID, commentID, replyID float64

	// ── Posts ─────────────────────────────────────────────────────────────────

	t.Run("create_post", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/feed", tokenA, map[string]string{
			"title":    "Test Post " + sfx,
			"content":  "Automated test post content",
			"category": "Test",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		postID = d["ID"].(float64)
		t.Logf("✅ Post created (id: %.0f)", postID)
	})

	t.Run("get_posts_list", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/feed?page=1&limit=5", tokenA, nil)
		assertStatus(t, 200, code, res)
		if res["data"] == nil {
			t.Fatal("missing data field in response")
		}
		t.Log("✅ Post list fetched (with counts)")
	})

	t.Run("update_post", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/feed/%.0f", postID), tokenA, map[string]string{
			"content": "Updated post content",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Post updated")
	})

	t.Run("cannot_update_others_post", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/feed/%.0f", postID), tokenB, map[string]string{
			"content": "should fail",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Cross-user post update blocked (403)")
	})

	// ── Comments ──────────────────────────────────────────────────────────────

	t.Run("add_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/comments", postID), tokenB, map[string]any{
			"content": "Great post!",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		commentID = d["ID"].(float64)
		t.Logf("✅ Comment added (id: %.0f)", commentID)
	})

	t.Run("reply_to_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/comments", postID), tokenA, map[string]any{
			"content":           "Thanks for the comment!",
			"parent_comment_id": commentID,
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		replyID = d["ID"].(float64)
		t.Logf("✅ Reply added (id: %.0f)", replyID)
	})

	t.Run("reply_to_reply", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/comments", postID), tokenB, map[string]any{
			"content":           "Deep nested reply!",
			"parent_comment_id": replyID,
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Reply-to-reply added (infinite nesting)")
	})

	t.Run("get_post_detail_with_nested_comments", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/feed/%.0f", postID), tokenA, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		comments := d["comments"].([]any)
		if len(comments) == 0 {
			t.Fatal("expected at least one top-level comment")
		}
		// Verify nesting
		topComment := comments[0].(map[string]any)
		replies := topComment["replies"].([]any)
		if len(replies) == 0 {
			t.Fatal("expected nested reply inside comment")
		}
		t.Logf("✅ Post detail: %d top-level comment(s) with %d reply(ies)", len(comments), len(replies))
	})

	t.Run("update_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, fmt.Sprintf("/api/comments/%.0f", commentID), tokenB, map[string]any{
			"content": "Updated comment",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Comment updated")
	})

	// ── Reactions ─────────────────────────────────────────────────────────────

	t.Run("react_to_post", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/react", postID), tokenB, map[string]any{
			"type": "like",
		})
		assertStatus(t, 200, code, res)
		if dataMap(res)["action"] != "added" {
			t.Fatal("expected action=added")
		}
		t.Log("✅ Reacted to post")
	})

	t.Run("toggle_reaction_removes_it", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/react", postID), tokenB, map[string]any{
			"type": "like",
		})
		assertStatus(t, 200, code, res)
		if dataMap(res)["action"] != "removed" {
			t.Fatal("expected action=removed")
		}
		t.Log("✅ Same reaction toggles off")
	})

	t.Run("change_reaction_updates_it", func(t *testing.T) {
		jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/react", postID), tokenB, map[string]any{"type": "like"})
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/react", postID), tokenB, map[string]any{"type": "love"})
		assertStatus(t, 200, code, res)
		if dataMap(res)["action"] != "updated" {
			t.Fatal("expected action=updated")
		}
		t.Log("✅ Different reaction replaces existing")
	})

	t.Run("react_to_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/comments/%.0f/react", commentID), tokenA, map[string]any{
			"type": "haha",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Reacted to comment")
	})

	t.Run("react_to_reply", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/comments/%.0f/react", replyID), tokenB, map[string]any{
			"type": "wow",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Reacted to reply")
	})

	t.Run("invalid_reaction_type", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/react", postID), tokenA, map[string]any{
			"type": "notanemoji",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid reaction type rejected")
	})

	// ── Votes ─────────────────────────────────────────────────────────────────

	t.Run("upvote_post", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/vote", postID), tokenB, map[string]any{"value": 1})
		assertStatus(t, 200, code, res)
		if dataMap(res)["action"] != "added" {
			t.Fatal("expected action=added")
		}
		t.Log("✅ Post upvoted")
	})

	t.Run("same_vote_removes_it", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/vote", postID), tokenB, map[string]any{"value": 1})
		assertStatus(t, 200, code, res)
		if dataMap(res)["action"] != "removed" {
			t.Fatal("expected action=removed")
		}
		t.Log("✅ Same vote toggles off")
	})

	t.Run("opposite_vote_flips_it", func(t *testing.T) {
		jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/vote", postID), tokenA, map[string]any{"value": 1})
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/vote", postID), tokenA, map[string]any{"value": -1})
		assertStatus(t, 200, code, res)
		if dataMap(res)["action"] != "flipped" {
			t.Fatal("expected action=flipped")
		}
		t.Log("✅ Opposite vote flips direction")
	})

	t.Run("vote_on_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/comments/%.0f/vote", commentID), tokenA, map[string]any{"value": 1})
		assertStatus(t, 200, code, res)
		t.Log("✅ Comment voted")
	})

	t.Run("invalid_vote_value", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%.0f/vote", postID), tokenA, map[string]any{"value": 0})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid vote value rejected")
	})

	// ── Cleanup ───────────────────────────────────────────────────────────────

	t.Run("delete_comment", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/comments/%.0f", commentID), tokenB, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Comment deleted")
	})

	t.Run("delete_post", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/feed/%.0f", postID), tokenA, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Post deleted")
	})
}
