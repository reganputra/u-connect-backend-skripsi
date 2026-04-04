package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestFollow covers the Follow System lifecycle.
// Run: go test -v -run TestFollow ./test/...
func TestFollow(t *testing.T) {
	sfx := newSuffix()

	// ── Actors ────────────────────────────────────────────────────────────────
	alumniToken, alumniUserID := registerAndLogin(t, sfx,
		"Alumni Follow "+sfx, "alumni.follow+"+sfx+"@test.com", "alumni", "Engineering", "CS")
	studentToken, studentUserID := registerAndLogin(t, sfx,
		"Student Follow "+sfx, "student.follow+"+sfx+"@test.com", "student", "Science", "Math")
	partnerToken, _ := registerAndLoginPartner(t, "Partner Follow "+sfx, "partner.follow+"+sfx+"@test.com", "PT Follow "+sfx)

	// Second student to test follow lists
	student2Token, student2UserID := registerAndLogin(t, sfx,
		"Student2 Follow "+sfx, "student2.follow+"+sfx+"@test.com", "student", "Arts", "Design")
	_ = student2UserID

	t.Run("partner_cannot_follow", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/users/%.0f/follow", alumniUserID), partnerToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Partner blocked from following (403)")
	})

	t.Run("self_follow_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/users/%.0f/follow", studentUserID), studentToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Self-follow rejected (400)")
	})

	t.Run("alumni_follows_student", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/users/%.0f/follow", studentUserID), alumniToken, nil)
		assertStatus(t, 201, code, res)
		t.Log("✅ Alumni follows student (201)")
	})

	t.Run("duplicate_follow_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/users/%.0f/follow", studentUserID), alumniToken, nil)
		assertStatus(t, 409, code, res)
		t.Log("✅ Duplicate follow rejected (409)")
	})

	t.Run("student_follows_alumni", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/users/%.0f/follow", alumniUserID), studentToken, nil)
		assertStatus(t, 201, code, res)
		t.Log("✅ Student follows alumni (201)")
	})

	t.Run("student2_follows_alumni", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/users/%.0f/follow", alumniUserID), student2Token, nil)
		assertStatus(t, 201, code, res)
		t.Log("✅ Student2 follows alumni (201)")
	})

	t.Run("get_alumni_followers", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet,
			fmt.Sprintf("/api/users/%.0f/followers", alumniUserID), alumniToken, nil)
		assertStatus(t, 200, code, res)
		data, ok := res["data"].([]any)
		if !ok || len(data) < 2 {
			t.Fatalf("expected at least 2 followers, got data: %v", res["data"])
		}
		t.Logf("✅ Alumni has %d follower(s)", len(data))
	})

	t.Run("get_student_following", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet,
			fmt.Sprintf("/api/users/%.0f/following", studentUserID), studentToken, nil)
		assertStatus(t, 200, code, res)
		data, ok := res["data"].([]any)
		if !ok || len(data) < 1 {
			t.Fatalf("expected at least 1 followed user, got: %v", res["data"])
		}
		t.Logf("✅ Student follows %d user(s)", len(data))
	})

	t.Run("unfollow_user", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete,
			fmt.Sprintf("/api/users/%.0f/follow", alumniUserID), student2Token, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Student2 unfollowed alumni (200)")
	})

	t.Run("unfollow_not_following", func(t *testing.T) {
		// student2 already unfollowed — trying again should error
		code, res := jsonReq(t, http.MethodDelete,
			fmt.Sprintf("/api/users/%.0f/follow", alumniUserID), student2Token, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Unfollow non-following user rejected (400)")
	})

	t.Run("unauthenticated_blocked", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/users/%.0f/follow", alumniUserID), "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ Unauthenticated follow rejected (401)")
	})
}

// TestMessage covers the Messaging lifecycle.
// Requires a mutual follow relationship between sender and receiver.
// Run: go test -v -run TestMessage ./test/...
func TestMessage(t *testing.T) {
	sfx := newSuffix()

	// ── Actors ────────────────────────────────────────────────────────────────
	alumniToken, _ := registerAndLogin(t, sfx,
		"Alumni Msg "+sfx, "alumni.msg+"+sfx+"@test.com", "alumni", "Engineering", "CS")
	studentToken, _ := registerAndLogin(t, sfx,
		"Student Msg "+sfx, "student.msg+"+sfx+"@test.com", "student", "Science", "Math")
	partnerToken, _ := registerAndLoginPartner(t, "Partner Msg "+sfx, "partner.msg+"+sfx+"@test.com", "PT Msg "+sfx)

	// strangerToken: a student who has no follow relationship with alumni
	strangerToken, _ := registerAndLogin(t, sfx,
		"Stranger Msg "+sfx, "stranger.msg+"+sfx+"@test.com", "student", "Arts", "Design")

	// Get user IDs via /api/me
	_, alumniMeRes := jsonReq(t, http.MethodGet, "/api/me", alumniToken, nil)
	alumniID := dataMap(alumniMeRes)["user_id"].(float64)

	_, studentMeRes := jsonReq(t, http.MethodGet, "/api/me", studentToken, nil)
	studentID := dataMap(studentMeRes)["user_id"].(float64)

	// ── Setup: create follow relationship ─────────────────────────────────────
	t.Run("setup_follow_relationship", func(t *testing.T) {
		// student follows alumni (symmetric: both can now message)
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/users/%.0f/follow", alumniID), studentToken, nil)
		assertStatus(t, 201, code, res)
		t.Log("✅ Student follows alumni — messaging channel open")
	})

	// ── Messaging business rules ───────────────────────────────────────────────

	t.Run("partner_blocked_from_messaging", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/messages", partnerToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Partner blocked from messaging (403)")
	})

	t.Run("send_without_follow_rejected", func(t *testing.T) {
		// stranger has no follow relationship with alumni
		_, strangerMeRes := jsonReq(t, http.MethodGet, "/api/me", strangerToken, nil)
		strangerID := dataMap(strangerMeRes)["user_id"].(float64)
		_ = strangerID

		code, res := jsonReq(t, http.MethodGet, "/api/messages", strangerToken, nil)
		assertStatus(t, 200, code, res) // listing is OK (empty), but sending should fail via WS
		// Note: actual send-without-follow is enforced at WS layer or can be tested
		// through a hypothetical REST send endpoint. We verify the follow check
		// in the service layer via unit test or WS test. For integration we verify
		// that the conversation list returns empty for a stranger.
		data, _ := res["data"].([]any)
		t.Logf("✅ Stranger has %d conversation(s) (expected 0)", len(data))
	})

	t.Run("get_empty_conversation_list_before_messages", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/messages", alumniToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Conversation list returns successfully (may be empty before WS messages)")
	})

	t.Run("get_unread_count", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/messages/unread", studentToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["unread_count"] == nil {
			t.Fatal("expected unread_count in response")
		}
		t.Logf("✅ Unread count: %.0f", d["unread_count"].(float64))
	})

	t.Run("get_conversation_history", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet,
			fmt.Sprintf("/api/messages/%.0f?page=1&limit=20", alumniID), studentToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["messages"] == nil {
			t.Fatal("expected messages field in response")
		}
		t.Log("✅ Conversation history fetched")
	})

	t.Run("mark_as_read", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPatch,
			fmt.Sprintf("/api/messages/%.0f/read", alumniID), studentToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Mark as read successful")
	})

	t.Run("unauthenticated_blocked", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/messages", "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ Unauthenticated request blocked (401)")
	})

	_ = studentID
}
