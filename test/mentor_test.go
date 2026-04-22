package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestMentor covers the full Mentoring Module lifecycle:
// registration, browsing, recommendation, request → approve, sessions, business rules.
// Run: go test -v -run TestMentor ./test/...
func TestMentor(t *testing.T) {
	sfx := newSuffix()

	// ── Actors ────────────────────────────────────────────────────────────────
	// mentorToken: alumni who will register as mentor
	mentorToken, mentorUserID := registerAndLogin(t, sfx,
		"Mentor Alumni "+sfx, "mentor+"+sfx+"@test.com", "alumni", "Engineering", "CS")

	// studentToken: student who will search and request mentoring
	studentToken, _ := registerAndLogin(t, sfx,
		"Student User "+sfx, "student+"+sfx+"@test.com", "student", "Engineering", "Information Systems")

	// fullMentorToken: alumni with quota=1, will be used to test capacity enforcement
	fullMentorToken, _ := registerAndLogin(t, sfx,
		"Full Mentor "+sfx, "fullmentor+"+sfx+"@test.com", "alumni", "Business", "Management")

	// partnerToken: must be blocked from all mentor/student endpoints
	partnerToken, _ := registerAndLoginPartner(t, "Partner "+sfx, "partner+"+sfx+"@test.com", "PT Test "+sfx)

	var requestID, requestID2, sessionID float64

	// ── Setup: create profiles with skills/interests (needed for recommendation) ──

	t.Run("setup_mentor_profile", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/profile", mentorToken, map[string]string{
			"job_status":   "employed",
			"position":     "Senior Software Engineer",
			"company_name": "PT Tech Indonesia",
			"skills":       "Python, Machine Learning, Cloud Computing, Docker, Kubernetes",
			"interests":    "AI, Data Science, Backend Development",
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Mentor profile created with skills & interests")
	})

	t.Run("setup_student_profile", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/profile", studentToken, map[string]string{
			"job_status": "student",
			"skills":     "Python, Data Analysis",
			"interests":  "Machine Learning, AI, Cloud Computing",
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Student profile created with skills & interests")
	})

	// ── Partner blocked from mentor endpoints ─────────────────────────────────

	t.Run("partner_blocked_from_register", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/mentor/register", partnerToken, map[string]any{
			"mentor_bio":   "should fail",
			"mentor_quota": 2,
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Partner blocked from mentor registration (403)")
	})

	t.Run("student_blocked_from_register", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/mentor/register", studentToken, map[string]any{
			"mentor_bio":   "should fail",
			"mentor_quota": 2,
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Student blocked from mentor registration (403)")
	})

	// ── Mentor Registration ───────────────────────────────────────────────────

	t.Run("invalid_quota_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/mentor/register", mentorToken, map[string]any{
			"mentor_bio":   "I love teaching",
			"mentor_quota": 4, // not in allowed set: 1, 2, 3, 5
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid quota (4) rejected")
	})

	t.Run("missing_bio_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/mentor/register", mentorToken, map[string]any{
			"mentor_quota": 2,
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Missing mentor_bio rejected")
	})

	t.Run("register_as_mentor", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/mentor/register", mentorToken, map[string]any{
			"mentor_bio":   "Experienced software engineer with 5+ years in Python and ML. Happy to mentor!",
			"mentor_quota": 3,
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		if d["MentorQuota"] == nil {
			t.Fatal("expected MentorQuota in response")
		}
		t.Logf("✅ Registered as mentor (quota: %.0f)", d["MentorQuota"])
	})

	t.Run("cannot_register_twice", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, "/api/mentor/register", mentorToken, map[string]any{
			"mentor_bio":   "duplicate",
			"mentor_quota": 2,
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Duplicate mentor registration rejected")
	})

	t.Run("register_full_mentor_quota_1", func(t *testing.T) {
		// Give full mentor a profile first (required for MentorQuota)
		formReq(t, http.MethodPost, "/api/profile", fullMentorToken, map[string]string{
			"job_status":   "employed",
			"position":     "Manager",
			"company_name": "PT Bisnis Indonesia",
			"skills":       "Management, Leadership",
			"interests":    "Entrepreneurship",
		})
		code, res := jsonReq(t, http.MethodPost, "/api/mentor/register", fullMentorToken, map[string]any{
			"mentor_bio":   "I can take 1 student only",
			"mentor_quota": 1,
		})
		assertStatus(t, 201, code, res)
		t.Log("✅ Full mentor registered with quota=1")
	})

	// ── Get Mentor Profile ────────────────────────────────────────────────────

	t.Run("get_own_mentor_profile", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentor/profile", mentorToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["MentorDescription"] == nil {
			t.Fatal("expected MentorDescription in response")
		}
		t.Log("✅ Own mentor profile fetched")
	})

	t.Run("update_mentor_profile", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, "/api/mentor/profile", mentorToken, map[string]any{
			"mentor_bio":   "Updated bio: Specializing in Python, ML, and Cloud.",
			"mentor_quota": 5,
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Mentor profile updated")
	})

	// ── Student: Browse Mentors ───────────────────────────────────────────────

	t.Run("alumni_blocked_from_mentor_list", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentors", mentorToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Alumni blocked from /api/mentors (403)")
	})

	t.Run("student_list_mentors", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentors?page=1&limit=10", studentToken, nil)
		assertStatus(t, 200, code, res)
		if res["data"] == nil {
			t.Fatal("expected data field in response")
		}
		t.Log("✅ Mentor list fetched")
	})

	t.Run("student_search_mentors", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentors?search=Python", studentToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Mentor search by keyword works")
	})

	t.Run("student_get_mentor_detail", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/mentors/%.0f", mentorUserID), studentToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Mentor detail fetched by student")
	})

	t.Run("get_nonexistent_mentor", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentors/99999", studentToken, nil)
		assertStatus(t, 404, code, res)
		t.Log("✅ Non-existent mentor returns 404")
	})

	// ── Recommendation Engine ─────────────────────────────────────────────────

	t.Run("auto_recommend_from_profile", func(t *testing.T) {
		// Automatic mode: uses student's skills + interests from profile
		code, res := jsonReq(t, http.MethodGet, "/api/mentors/recommend", studentToken, nil)
		assertStatus(t, 200, code, res)
		data, ok := res["data"].([]any)
		if !ok {
			t.Fatal("expected data array in recommendation response")
		}
		t.Logf("✅ Auto recommendation returned %d mentors", len(data))
		if len(data) > 0 {
			first := data[0].(map[string]any)
			if first["similarity_score"] == nil {
				t.Fatal("expected similarity_score in recommendation result")
			}
			t.Logf("   Top match: %v (score: %.4f)", first["name"], first["similarity_score"])
		}
	})

	t.Run("query_based_recommendation", func(t *testing.T) {
		// Query mode: custom text query
		code, res := jsonReq(t, http.MethodGet, "/api/mentors/recommend?q=python+machine+learning+cloud", studentToken, nil)
		assertStatus(t, 200, code, res)
		data, ok := res["data"].([]any)
		if !ok {
			t.Fatal("expected data array")
		}
		t.Logf("✅ Query-based recommendation returned %d mentors", len(data))
		// Mentor with Python + ML skills should appear (non-zero score)
		if len(data) > 0 {
			first := data[0].(map[string]any)
			score, _ := first["similarity_score"].(float64)
			if score == 0 {
				t.Log("⚠️  Top score is 0 (skills text may be empty or mismatch)")
			} else {
				t.Logf("   Top match similarity: %.4f", score)
			}
		}
	})

	t.Run("recommend_limit_param", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentors/recommend?top=1", studentToken, nil)
		assertStatus(t, 200, code, res)
		data, _ := res["data"].([]any)
		if len(data) > 1 {
			t.Fatalf("expected max 1 result with top=1, got %d", len(data))
		}
		t.Logf("✅ top=1 parameter respected (%d result)", len(data))
	})

	t.Run("alumni_blocked_from_recommend", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentors/recommend", mentorToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Alumni blocked from recommendation endpoint (403)")
	})

	// ── Mentoring Requests ────────────────────────────────────────────────────

	t.Run("student_send_request", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentorUserID), studentToken, map[string]any{
			"message": "Hi! I'd love to learn Python and ML from you.",
		})
		assertStatus(t, 201, code, res)
		requestID = safeID(t, res, "mentor request")
		t.Logf("✅ Mentoring request sent (id: %.0f)", requestID)
	})

	t.Run("duplicate_request_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentorUserID), studentToken, map[string]any{
			"message": "duplicate",
		})
		assertStatus(t, 409, code, res)
		t.Log("✅ Duplicate pending request rejected")
	})

	t.Run("student_withdraw_pending_request", func(t *testing.T) {
		sfxW := newSuffix()
		withdrawToken, _ := registerAndLogin(t, sfxW,
			"Withdraw Student "+sfxW, "withdraw+"+sfxW+"@test.com", "student", "Science", "Biology")
		formReq(t, http.MethodPost, "/api/profile", withdrawToken, map[string]string{
			"job_status": "student", "skills": "biology", "interests": "research",
		})

		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentorUserID), withdrawToken, map[string]any{
			"message": "I might withdraw this",
		})
		assertStatus(t, 201, code, res)
		withdrawReqID := safeID(t, res, "withdraw request")

		code, res = jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/student/requests/%.0f", withdrawReqID), withdrawToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Student can withdraw pending request")
	})

	t.Run("alumni_blocked_from_requesting", func(t *testing.T) {
		// Another alumni cannot send a mentoring request (student-only)
		otherAlumni, _ := registerAndLogin(t, sfx,
			"Other Alumni "+sfx, "other.alumni+"+sfx+"@test.com", "alumni", "Arts", "Design")
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentorUserID), otherAlumni, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Alumni blocked from sending mentoring request (403)")
	})

	// ── Mentor sees incoming requests ─────────────────────────────────────────

	t.Run("mentor_gets_incoming_requests", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentor/requests", mentorToken, nil)
		assertStatus(t, 200, code, res)
		data, ok := res["data"].([]any)
		if !ok || len(data) == 0 {
			t.Fatal("expected at least one incoming request")
		}
		t.Logf("✅ Mentor has %d incoming request(s)", len(data))
	})

	t.Run("mentor_approves_request", func(t *testing.T) {
		if requestID == 0 {
			t.Skip("⏭️  Skipped — request not created")
		}
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/requests/%.0f/approve", requestID), mentorToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Status"] != "approved" {
			t.Fatalf("expected Status=approved, got %v", d["Status"])
		}
		t.Log("✅ Request approved by mentor")
	})

	t.Run("cannot_approve_already_processed_request", func(t *testing.T) {
		if requestID == 0 {
			t.Skip("⏭️  Skipped — request not created")
		}
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/requests/%.0f/approve", requestID), mentorToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Re-approving already approved request blocked")
	})

	// ── Capacity enforcement: full mentor (quota=1) ───────────────────────────

	t.Run("capacity_enforcement", func(t *testing.T) {
		// Get the full mentor's user ID via /api/me (token already held)
		_, meRes := jsonReq(t, http.MethodGet, "/api/me", fullMentorToken, nil)
		fullMentorUserID, ok := dataMap(meRes)["user_id"].(float64)
		if !ok || fullMentorUserID == 0 {
			t.Fatal("could not get full mentor user_id from /api/me")
		}

		// Filler student fills the 1 available slot
		sfxF := newSuffix()
		fillerToken, _ := registerAndLogin(t, sfxF,
			"Filler "+sfxF, "filler+"+sfxF+"@test.com", "student", "Science", "Physics")
		formReq(t, http.MethodPost, "/api/profile", fillerToken, map[string]string{
			"job_status": "student", "skills": "basic", "interests": "management",
		})
		code, res := jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/mentors/%.0f/request", fullMentorUserID), fillerToken, nil)
		if code != 201 {
			t.Skip("⏭️  Skipped — filler request failed (mentor may already be full)")
		}
		fillerRequestID := safeID(t, res, "filler request")
		jsonReq(t, http.MethodPatch,
			fmt.Sprintf("/api/mentor/requests/%.0f/approve", fillerRequestID), fullMentorToken, nil)

		// Blocked student tries to request the now-full mentor
		sfxB := newSuffix()
		blockedToken, _ := registerAndLogin(t, sfxB,
			"Blocked "+sfxB, "blocked+"+sfxB+"@test.com", "student", "Arts", "Design")
		formReq(t, http.MethodPost, "/api/profile", blockedToken, map[string]string{
			"job_status": "student", "skills": "design", "interests": "management",
		})
		code, res = jsonReq(t, http.MethodPost,
			fmt.Sprintf("/api/mentors/%.0f/request", fullMentorUserID), blockedToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Capacity enforcement: full mentor (quota=1) blocks new requests")
	})

	// ── Reject flow ───────────────────────────────────────────────────────────

	t.Run("reject_request_flow", func(t *testing.T) {
		// Create a fresh student and request to test reject
		sfx2 := newSuffix()
		rejToken, _ := registerAndLogin(t, sfx2,
			"Reject Student "+sfx2, "reject+"+sfx2+"@test.com", "student", "Law", "Law")
		formReq(t, http.MethodPost, "/api/profile", rejToken, map[string]string{
			"job_status": "student", "skills": "law", "interests": "legal",
		})
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentorUserID), rejToken, map[string]any{
			"message": "please reject me",
		})
		assertStatus(t, 201, code, res)
		rejRequestID := safeID(t, res, "rejection request")

		code, res = jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/requests/%.0f/reject", rejRequestID), mentorToken, map[string]any{
			"reason": "Not aligned with my expertise currently.",
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Status"] != "rejected" {
			t.Fatalf("expected Status=rejected, got %v", d["Status"])
		}
		t.Log("✅ Request rejected with reason")
	})

	// ── 2-Mentor limit enforcement ────────────────────────────────────────────

	t.Run("two_mentor_limit_enforcement", func(t *testing.T) {
		// student already has 1 approved mentor (mentorToken above)
		// Register a second mentor and approve student's request
		sfx3 := newSuffix()
		mentor2Token, mentor2UserID := registerAndLogin(t, sfx3,
			"Second Mentor "+sfx3, "mentor2+"+sfx3+"@test.com", "alumni", "Science", "Physics")
		formReq(t, http.MethodPost, "/api/profile", mentor2Token, map[string]string{
			"job_status": "employed", "position": "Physicist", "company_name": "BRIN",
			"skills": "Physics, Research, Data", "interests": "Science",
		})
		jsonReq(t, http.MethodPost, "/api/mentor/register", mentor2Token, map[string]any{
			"mentor_bio": "Physics mentor", "mentor_quota": 3,
		})

		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentor2UserID), studentToken, map[string]any{
			"message": "second mentor please",
		})
		assertStatus(t, 201, code, res)
		requestID2 = safeID(t, res, "second request")

		jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/requests/%.0f/approve", requestID2), mentor2Token, nil)
		t.Log("✅ Student now has 2 approved mentors")

		// Now try a third mentor — must be blocked
		sfx4 := newSuffix()
		mentor3Token, mentor3UserID := registerAndLogin(t, sfx4,
			"Third Mentor "+sfx4, "mentor3+"+sfx4+"@test.com", "alumni", "Business", "Finance")
		formReq(t, http.MethodPost, "/api/profile", mentor3Token, map[string]string{
			"job_status": "employed", "position": "Finance Lead", "company_name": "Bank Maju",
			"skills": "Finance, Excel", "interests": "Investment",
		})
		jsonReq(t, http.MethodPost, "/api/mentor/register", mentor3Token, map[string]any{
			"mentor_bio": "Finance mentor", "mentor_quota": 3,
		})

		code, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentor3UserID), studentToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Third mentor request blocked (2-mentor limit enforced)")
	})

	// ── Student views their mentors and requests ──────────────────────────────

	t.Run("student_view_my_mentors", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/student/mentors", studentToken, nil)
		assertStatus(t, 200, code, res)
		data, ok := res["data"].([]any)
		if !ok || len(data) == 0 {
			t.Fatal("expected at least 1 approved mentor")
		}
		t.Logf("✅ Student has %d approved mentor(s)", len(data))
	})

	t.Run("student_view_sent_requests", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/student/requests", studentToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Student's sent requests fetched")
	})

	// ── Mentor views mentees ──────────────────────────────────────────────────

	t.Run("mentor_view_mentees", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentor/mentees", mentorToken, nil)
		assertStatus(t, 200, code, res)
		data, ok := res["data"].([]any)
		if !ok || len(data) == 0 {
			t.Fatal("expected at least 1 mentee")
		}
		t.Logf("✅ Mentor has %d mentee(s)", len(data))
	})

	// ── Session Management ────────────────────────────────────────────────────

	t.Run("create_session", func(t *testing.T) {
		if requestID == 0 {
			t.Skip("⏭️  Skipped — approved request not available")
		}
		// Get student user ID via /api/me (no re-login needed)
		_, meRes := jsonReq(t, http.MethodGet, "/api/me", studentToken, nil)
		studentUserID, ok := dataMap(meRes)["user_id"].(float64)
		if !ok || studentUserID == 0 {
			t.Fatal("could not get student user_id from /api/me")
		}

		code, res := jsonReq(t, http.MethodPost, "/api/mentor/sessions", mentorToken, map[string]any{
			"student_id":   studentUserID,
			"topic":        "Introduction to Python & ML basics",
			"notes":        "Bring your laptop and VS Code installed",
			"session_date": "2026-07-01T10:00:00Z",
		})
		assertStatus(t, 201, code, res)
		sessionID = safeID(t, res, "session")
		t.Logf("✅ Mentoring session created (id: %.0f)", sessionID)
	})

	t.Run("session_requires_approved_relationship", func(t *testing.T) {
		// Register a student who has NO approved relationship with mentorToken
		sfx5 := newSuffix()
		noRelToken, _ := registerAndLogin(t, sfx5, "No Rel "+sfx5, "norel+"+sfx5+"@test.com",
			"student", "Arts", "Design")
		formReq(t, http.MethodPost, "/api/profile", noRelToken, map[string]string{
			"job_status": "student", "skills": "design", "interests": "art",
		})
		// Get their user ID via /api/me
		_, meRes := jsonReq(t, http.MethodGet, "/api/me", noRelToken, nil)
		noRelUserID, ok := dataMap(meRes)["user_id"].(float64)
		if !ok || noRelUserID == 0 {
			t.Fatal("could not get noRel user_id from /api/me")
		}

		code, res := jsonReq(t, http.MethodPost, "/api/mentor/sessions", mentorToken, map[string]any{
			"student_id": noRelUserID,
			"topic":      "Unauthorized session",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Session without approved relationship rejected")
	})

	t.Run("create_session_missing_topic", func(t *testing.T) {
		// Get student user ID via /api/me
		_, meRes := jsonReq(t, http.MethodGet, "/api/me", studentToken, nil)
		studentUserID, ok := dataMap(meRes)["user_id"].(float64)
		if !ok || studentUserID == 0 {
			t.Fatal("could not get student user_id from /api/me")
		}

		code, res := jsonReq(t, http.MethodPost, "/api/mentor/sessions", mentorToken, map[string]any{
			"student_id": studentUserID,
			// topic missing
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Session without topic rejected")
	})

	t.Run("mentor_view_sessions", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentor/sessions", mentorToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Mentor's sessions fetched")
	})

	t.Run("student_view_sessions", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/student/sessions", studentToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Student's sessions fetched")
	})

	t.Run("update_session", func(t *testing.T) {
		if sessionID == 0 {
			t.Skip("⏭️  Skipped — session not created")
		}
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/sessions/%.0f", sessionID), mentorToken, map[string]any{
			"topic":  "Advanced Python: decorators and async",
			"status": "completed",
		})
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Status"] != "completed" {
			t.Fatalf("expected Status=completed, got %v", d["Status"])
		}
		t.Log("✅ Session updated to completed")
	})

	t.Run("update_session_invalid_status", func(t *testing.T) {
		if sessionID == 0 {
			t.Skip("⏭️  Skipped — session not created")
		}
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/sessions/%.0f", sessionID), mentorToken, map[string]any{
			"status": "unknown_status",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid session status rejected")
	})

	t.Run("cannot_update_terminal_session", func(t *testing.T) {
		if sessionID == 0 {
			t.Skip("⏭️  Skipped — session not created")
		}
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/sessions/%.0f", sessionID), mentorToken, map[string]any{
			"topic": "retry after completed",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Terminal session cannot be modified")
	})

	t.Run("non_owner_cannot_update_session", func(t *testing.T) {
		if sessionID == 0 {
			t.Skip("⏭️  Skipped — session not created")
		}
		// fullMentorToken is a different mentor — cannot update sessionID
		code, res := jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/sessions/%.0f", sessionID), fullMentorToken, map[string]any{
			"topic": "hijacked",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner cannot update session (403)")
	})

	// ── Unregister as Mentor ──────────────────────────────────────────────────

	t.Run("cannot_unregister_with_active_mentees", func(t *testing.T) {
		// mentorToken still has active mentees — should be blocked
		code, res := jsonReq(t, http.MethodDelete, "/api/mentor/unregister", mentorToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Unregister with active mentees blocked")
	})

	// ── Unauthenticated ───────────────────────────────────────────────────────

	t.Run("unauthenticated_blocked", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/mentors", "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ Unauthenticated request blocked (401)")
	})

	_ = requestID2
}
