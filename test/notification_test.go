package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	websocket "github.com/fasthttp/websocket"
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/scheduler"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
	"gorm.io/gorm"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

func getUnreadCount(t *testing.T, token string) float64 {
	status, res := jsonReq(t, http.MethodGet, "/api/notifications/unread", token, nil)
	assertStatus(t, http.StatusOK, status, res)
	return dataMap(res)["unread_count"].(float64)
}

func getNotifications(t *testing.T, token string) []any {
	status, res := jsonReq(t, http.MethodGet, "/api/notifications", token, nil)
	assertStatus(t, http.StatusOK, status, res)
	return dataMap(res)["notifications"].([]any)
}

func pickAny(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v
		}
	}
	return nil
}

func latestNotification(t *testing.T, token string) map[string]any {
	t.Helper()
	notifs := getNotifications(t, token)
	if len(notifs) == 0 {
		t.Fatal("No notifications returned")
	}
	notif, ok := notifs[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected notification payload: %T", notifs[0])
	}
	return notif
}

func notificationType(notif map[string]any) string {
	if v := pickAny(notif, "notification_type", "NotificationType"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func notificationRefID(notif map[string]any) float64 {
	if v := pickAny(notif, "reference_id", "ReferenceID"); v != nil {
		if id, ok := v.(float64); ok {
			return id
		}
	}
	return 0
}

func notificationBody(notif map[string]any) string {
	if v := pickAny(notif, "body", "Body"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func notificationReadCountDelta(t *testing.T, token string, before float64, delta float64) float64 {
	t.Helper()
	after := getUnreadCount(t, token)
	if after != before+delta {
		t.Fatalf("expected unread delta %v, baseline %v, got %v", delta, before, after)
	}
	return after
}

func openNotificationWS(t *testing.T, token string) *websocket.Conn {
	t.Helper()
	wsURL := "ws://localhost:8080/api/ws?token=" + url.QueryEscape(token)
	requestHeader := http.Header{}
	requestHeader.Set("Origin", "http://localhost:8080")
	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = 5 * time.Second
	conn, _, err := dialer.Dial(wsURL, requestHeader)
	if err != nil {
		t.Fatalf("failed to open websocket: %v", err)
	}
	return conn
}

func readWSPayload(t *testing.T, conn *websocket.Conn, timeout time.Duration) map[string]any {
	t.Helper()
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		t.Fatalf("failed to set websocket read deadline: %v", err)
	}
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("failed to decode websocket payload: %v", err)
	}
	return payload
}

type noopDeliverer struct{}

func (noopDeliverer) SendToUser(uint, []byte) bool { return false }

var testDBOnce sync.Once

func ensureTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	testDBOnce.Do(func() {
		utils.LoadEnvFile("../.env")
		config.ConnectDB()
	})
	if config.DB == nil {
		t.Fatal("test database is not initialized")
	}
	return config.DB
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestNotificationFlows(t *testing.T) {
	// 1. Setup users
	s := newSuffix()
	tokenUserA, idA := registerAndLogin(t, s, "User A Notif", "notifa_"+s+"@test.com", "alumni", "FIK", "IF")
	tokenUserB, idB := registerAndLogin(t, s, "User B Notif", "notifb_"+s+"@test.com", "student", "FIK", "IF")

	// ========================================================================
	// Scenario: follow_creates_notification
	// ========================================================================
	t.Run("Follow Creates Notification", func(t *testing.T) {
		baseUnread := getUnreadCount(t, tokenUserA)

		status, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/users/%.0f/follow", idA), tokenUserB, nil)
		assertStatus(t, http.StatusCreated, status, res)

		// Check User A got a notification
		time.Sleep(100 * time.Millisecond) // Give async ops/DB a tiny moment just in case
		countA := getUnreadCount(t, tokenUserA)
		if countA != baseUnread+1 {
			t.Errorf("Expected unread to increase by 1 for User A, was %v now %v", baseUnread, countA)
		}

		notifs := getNotifications(t, tokenUserA)
		if len(notifs) == 0 {
			t.Fatal("No notifications returned")
		}
		n1 := notifs[0].(map[string]any)
		nType := pickAny(n1, "notification_type", "NotificationType")
		if nType != "new_follower" {
			t.Errorf("Expected type new_follower, got %v", nType)
		}
		refID := pickAny(n1, "reference_id", "ReferenceID")
		if ref, ok := refID.(float64); !ok || int(ref) != int(idB) {
			t.Errorf("Expected refID UserB(%v), got %v", idB, refID)
		}
	})

	// ========================================================================
	// Scenario: comment_creates_notification & comment_throttle
	// ========================================================================
	t.Run("Feed Comment Throttle", func(t *testing.T) {
		baseUnread := getUnreadCount(t, tokenUserA)

		// A creates post
		status, resA := formReq(t, http.MethodPost, "/api/feed", tokenUserA, map[string]string{
			"title":   "Notif Post " + s,
			"content": "Hello Notif",
		})
		assertStatus(t, http.StatusCreated, status, resA)
		postID := safeID(t, resA, "post")

		// B comments on A's post
		status, _ = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%v/comments", postID), tokenUserB, map[string]string{
			"content": "First Comment",
		})
		assertStatus(t, http.StatusCreated, status, nil)

		// Assert A got 1 notification
		countA := getUnreadCount(t, tokenUserA)
		if countA != baseUnread+1 {
			t.Errorf("Expected unread to increase by 1 (first comment), was %v now %v", baseUnread, countA)
		}

		// B comments AGAIN on same post within throttle window
		status, _ = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%v/comments", postID), tokenUserB, map[string]string{
			"content": "Second Comment (Throttled)",
		})
		assertStatus(t, http.StatusCreated, status, nil)

		// Assert A still has 1 notification
		countA2 := getUnreadCount(t, tokenUserA)
		if countA2 != baseUnread+1 {
			t.Errorf("Expected unread unchanged after throttled comment, got %v (baseline %v)", countA2, baseUnread)
		}
	})

	// ========================================================================
	// Scenario: post reaction notification
	// ========================================================================
	t.Run("Feed Reaction Notification", func(t *testing.T) {
		s3 := newSuffix()
		ownerToken, _ := registerAndLogin(t, s3, "Reaction Owner", "react-owner+"+s3+"@test.com", "alumni", "FIK", "IF")
		reactorToken, _ := registerAndLogin(t, s3, "Reaction Reactor", "react-reactor+"+s3+"@test.com", "student", "FIK", "IF")

		baseUnread := getUnreadCount(t, ownerToken)

		status, res := formReq(t, http.MethodPost, "/api/feed", ownerToken, map[string]string{
			"title":   "Reaction Post " + s3,
			"content": "Post to be reacted to",
		})
		assertStatus(t, http.StatusCreated, status, res)
		postID := safeID(t, res, "post")

		status, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/feed/%v/react", postID), reactorToken, map[string]string{
			"type": "like",
		})
		assertStatus(t, http.StatusOK, status, res)
		if dataMap(res)["action"] != "added" {
			t.Fatalf("expected reaction action=added, got %v", dataMap(res)["action"])
		}

		notificationReadCountDelta(t, ownerToken, baseUnread, 1)
		notif := latestNotification(t, ownerToken)
		if got := notificationType(notif); got != "post_reacted" {
			t.Fatalf("expected post_reacted notification, got %v", got)
		}
		if got := notificationRefID(notif); int(got) != int(postID) {
			t.Fatalf("expected ref id %.0f, got %v", postID, got)
		}
	})

	// ========================================================================
	// Scenario: group kick notification
	// ========================================================================
	t.Run("Group Kick Notification", func(t *testing.T) {
		s4 := newSuffix()
		ownerToken, _ := registerAndLogin(t, s4, "Group Owner Notif", "group-owner+"+s4+"@test.com", "alumni", "FIK", "IF")
		memberToken, memberID := registerAndLogin(t, s4, "Group Member Notif", "group-member+"+s4+"@test.com", "student", "FIK", "IF")

		status, res := formReq(t, http.MethodPost, "/api/groups", ownerToken, map[string]string{
			"title":    "Notif Group " + s4,
			"category": "Technology",
			"rules":    "Be respectful",
		})
		assertStatus(t, http.StatusCreated, status, res)
		groupID := safeID(t, res, "group")

		status, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/groups/%.0f/join", groupID), memberToken, nil)
		assertStatus(t, http.StatusOK, status, res)

		baseUnread := getUnreadCount(t, memberToken)
		status, res = jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/groups/%.0f/members/%.0f", groupID, memberID), ownerToken, map[string]any{
			"reason": "spam posting",
		})
		assertStatus(t, http.StatusOK, status, res)

		notificationReadCountDelta(t, memberToken, baseUnread, 1)
		notif := latestNotification(t, memberToken)
		if got := notificationType(notif); got != "group_kicked" {
			t.Fatalf("expected group_kicked notification, got %v", got)
		}
		if got := notificationRefID(notif); int(got) != int(groupID) {
			t.Fatalf("expected ref id %.0f, got %v", groupID, got)
		}
		if !strings.Contains(notificationBody(notif), "spam posting") {
			t.Fatalf("expected kick reason in body, got %q", notificationBody(notif))
		}
	})

	// ========================================================================
	// Scenario: job application status notification
	// ========================================================================
	t.Run("Job Application Status Notification", func(t *testing.T) {
		s5 := newSuffix()
		ownerToken, _ := registerAndLogin(t, s5, "Job Owner Notif", "job-owner+"+s5+"@test.com", "alumni", "FIK", "IF")
		applicantToken, _ := registerAndLogin(t, s5, "Job Applicant Notif", "job-applicant+"+s5+"@test.com", "student", "FIK", "IF")

		status, res := formReq(t, http.MethodPost, "/api/jobs", ownerToken, map[string]string{
			"title":        "Notif Job " + s5,
			"description":  "A job for notification testing",
			"company_name": "PT Notif Corp",
			"location":     "Jakarta",
			"job_type":     "full-time",
			"status":       "open",
		})
		assertStatus(t, http.StatusCreated, status, res)
		jobID := safeID(t, res, "job")

		status, res = formReq(t, http.MethodPost, fmt.Sprintf("/api/jobs/%.0f/apply", jobID), applicantToken, map[string]string{
			"cover_letter": "I would like to apply",
			"resume_url":   "https://example.com/resume.pdf",
		})
		assertStatus(t, http.StatusCreated, status, res)
		applicationID := safeID(t, res, "application")

		baseUnread := getUnreadCount(t, applicantToken)
		status, res = jsonReq(t, http.MethodPut, fmt.Sprintf("/api/jobs/applications/%.0f/status", applicationID), ownerToken, map[string]any{
			"status": "reviewed",
		})
		assertStatus(t, http.StatusOK, status, res)

		notificationReadCountDelta(t, applicantToken, baseUnread, 1)
		notif := latestNotification(t, applicantToken)
		if got := notificationType(notif); got != "job_application_updated" {
			t.Fatalf("expected job_application_updated notification, got %v", got)
		}
		if got := notificationRefID(notif); int(got) != int(applicationID) {
			t.Fatalf("expected ref id %.0f, got %v", applicationID, got)
		}
	})

	// ========================================================================
	// Scenario: mentor notifications
	// ========================================================================
	t.Run("Mentor Notification Flows", func(t *testing.T) {
		mentorToken, mentorUserID := registerAndLogin(t, newSuffix(), "Mentor Notif", "mentor-notif@test.com", "alumni", "FIK", "IF")
		studentToken, studentUserID := registerAndLogin(t, newSuffix(), "Student Notif", "student-notif@test.com", "student", "FIK", "IF")

		status, res := formReq(t, http.MethodPost, "/api/profile", mentorToken, map[string]string{
			"job_status":   "employed",
			"position":     "Engineer",
			"company_name": "PT Mentor Co",
			"skills":       "Go, Cloud, Architecture",
			"interests":    "System Design, Mentoring",
		})
		assertStatus(t, http.StatusCreated, status, res)

		status, res = jsonReq(t, http.MethodPost, "/api/mentor/register", mentorToken, map[string]any{
			"mentor_bio":   "Happy to mentor students",
			"mentor_quota": 2,
		})
		assertStatus(t, http.StatusCreated, status, res)

		mentorBaseUnread := getUnreadCount(t, mentorToken)
		status, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentorUserID), studentToken, map[string]any{
			"message": "Please mentor me",
		})
		assertStatus(t, http.StatusCreated, status, res)
		requestID := safeID(t, res, "mentor request")

		notificationReadCountDelta(t, mentorToken, mentorBaseUnread, 1)
		notif := latestNotification(t, mentorToken)
		if got := notificationType(notif); got != "mentor_request_received" {
			t.Fatalf("expected mentor_request_received notification, got %v", got)
		}
		if got := notificationRefID(notif); int(got) != int(requestID) {
			t.Fatalf("expected ref id %.0f, got %v", requestID, got)
		}

		studentBaseUnread := getUnreadCount(t, studentToken)
		status, res = jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/requests/%.0f/approve", requestID), mentorToken, nil)
		assertStatus(t, http.StatusOK, status, res)

		notificationReadCountDelta(t, studentToken, studentBaseUnread, 1)
		notif = latestNotification(t, studentToken)
		if got := notificationType(notif); got != "mentor_request_approved" {
			t.Fatalf("expected mentor_request_approved notification, got %v", got)
		}

		status, res = jsonReq(t, http.MethodPost, "/api/mentor/sessions", mentorToken, map[string]any{
			"student_id":   studentUserID,
			"topic":        "Architecture review",
			"session_date": "2026-06-01T09:00:00Z",
		})
		assertStatus(t, http.StatusCreated, status, res)
		sessionID := safeID(t, res, "session")

		notificationReadCountDelta(t, studentToken, studentBaseUnread+1, 1)
		notif = latestNotification(t, studentToken)
		if got := notificationType(notif); got != "new_session" {
			t.Fatalf("expected new_session notification, got %v", got)
		}
		if got := notificationRefID(notif); int(got) != int(sessionID) {
			t.Fatalf("expected ref id %.0f, got %v", sessionID, got)
		}
	})

	t.Run("Mentor Request Rejected Notification", func(t *testing.T) {
		mentorToken, mentorUserID := registerAndLogin(t, newSuffix(), "Mentor Reject", "mentor-reject@test.com", "alumni", "FIK", "IF")
		studentToken, _ := registerAndLogin(t, newSuffix(), "Student Reject", "student-reject@test.com", "student", "FIK", "IF")

		status, res := formReq(t, http.MethodPost, "/api/profile", mentorToken, map[string]string{
			"job_status":   "employed",
			"position":     "Engineer",
			"company_name": "PT Mentor Reject",
			"skills":       "Go",
			"interests":    "Backend",
		})
		assertStatus(t, http.StatusCreated, status, res)

		status, res = jsonReq(t, http.MethodPost, "/api/mentor/register", mentorToken, map[string]any{
			"mentor_bio":   "Busy mentor",
			"mentor_quota": 1,
		})
		assertStatus(t, http.StatusCreated, status, res)

		status, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/mentors/%.0f/request", mentorUserID), studentToken, map[string]any{
			"message": "Please reject this one",
		})
		assertStatus(t, http.StatusCreated, status, res)
		requestID := safeID(t, res, "mentor request to reject")

		baseUnread := getUnreadCount(t, studentToken)
		status, res = jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/mentor/requests/%.0f/reject", requestID), mentorToken, map[string]any{
			"reason": "Not available",
		})
		assertStatus(t, http.StatusOK, status, res)

		notificationReadCountDelta(t, studentToken, baseUnread, 1)
		notif := latestNotification(t, studentToken)
		if got := notificationType(notif); got != "mentor_request_rejected" {
			t.Fatalf("expected mentor_request_rejected notification, got %v", got)
		}
	})

	// ========================================================================
	// Scenario: report rejection notification
	// ========================================================================
	t.Run("Report Rejection Notification", func(t *testing.T) {
		adminToken := loginAdmin(t)
		reporterToken, _ := registerAndLogin(t, newSuffix(), "Report Reporter", "report-reporter@test.com", "alumni", "FIK", "IF")
		authorToken, _ := registerAndLogin(t, newSuffix(), "Report Author", "report-author@test.com", "alumni", "FIK", "IF")

		status, res := formReq(t, http.MethodPost, "/api/feed", authorToken, map[string]string{
			"title":   "Reportable Post",
			"content": "This will be reported",
		})
		assertStatus(t, http.StatusCreated, status, res)
		postID := safeID(t, res, "reportable post")

		status, res = jsonReq(t, http.MethodPost, "/api/reports", reporterToken, map[string]any{
			"target_type": "post",
			"target_id":   postID,
			"report_type": "spam",
		})
		assertStatus(t, http.StatusCreated, status, res)
		reportID := safeID(t, res, "report")

		baseUnread := getUnreadCount(t, reporterToken)
		status, res = jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/admin/reports/%.0f/reject", reportID), adminToken, map[string]any{
			"admin_note": "Does not violate policy",
		})
		assertStatus(t, http.StatusOK, status, res)

		notificationReadCountDelta(t, reporterToken, baseUnread, 1)
		notif := latestNotification(t, reporterToken)
		if got := notificationType(notif); got != "report_rejected" {
			t.Fatalf("expected report_rejected notification, got %v", got)
		}
		if got := notificationRefID(notif); int(got) != int(reportID) {
			t.Fatalf("expected ref id %.0f, got %v", reportID, got)
		}
	})

	// ========================================================================
	// Scenario: websocket message notification
	// ========================================================================
	t.Run("WebSocket Message Notification", func(t *testing.T) {
		s6 := newSuffix()
		senderToken, _ := registerAndLogin(t, s6, "WS Sender", "ws-sender+"+s6+"@test.com", "alumni", "FIK", "IF")
		receiverToken, receiverID := registerAndLogin(t, s6, "WS Receiver", "ws-receiver+"+s6+"@test.com", "student", "FIK", "IF")

		status, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/users/%.0f/follow", receiverID), senderToken, nil)
		assertStatus(t, http.StatusCreated, status, res)

		receiverBaseUnread := getUnreadCount(t, receiverToken)
		receiverConn := openNotificationWS(t, receiverToken)
		defer receiverConn.Close()
		senderConn := openNotificationWS(t, senderToken)
		defer senderConn.Close()

		if err := senderConn.WriteJSON(map[string]any{
			"receiver_id": receiverID,
			"content":     "Hello from websocket",
		}); err != nil {
			t.Fatalf("failed to send websocket message: %v", err)
		}

		var gotNotification map[string]any
		for i := 0; i < 3; i++ {
			payload := readWSPayload(t, receiverConn, 5*time.Second)
			if payload["type"] == "notification" {
				if data, ok := payload["data"].(map[string]any); ok {
					gotNotification = data
					break
				}
			}
		}
		if gotNotification == nil {
			t.Fatal("did not receive notification payload over websocket")
		}
		if got := notificationType(gotNotification); got != "new_message" {
			t.Fatalf("expected new_message notification, got %v", got)
		}

		notificationReadCountDelta(t, receiverToken, receiverBaseUnread, 1)
	})

	t.Run("Event Reminder Notification", func(t *testing.T) {
		db := ensureTestDB(t)
		notifSvc := service.NewNotificationService(repository.NewNotificationRepository(db), noopDeliverer{})

		s7 := newSuffix()
		ownerToken, _ := registerAndLogin(t, s7, "Reminder Owner", "reminder-owner+"+s7+"@test.com", "alumni", "FIK", "IF")
		participantToken, _ := registerAndLogin(t, s7, "Reminder Participant", "reminder-participant+"+s7+"@test.com", "student", "FIK", "IF")

		startTime := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
		status, res := formReq(t, http.MethodPost, "/api/events", ownerToken, map[string]string{
			"title":      "Reminder Event " + s7,
			"status":     "upcoming",
			"start_time": startTime,
		})
		assertStatus(t, http.StatusCreated, status, res)
		eventID := safeID(t, res, "reminder event")

		status, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/register", eventID), participantToken, nil)
		assertStatus(t, http.StatusOK, status, res)

		baseUnread := getUnreadCount(t, participantToken)
		scheduler.RunEventRemindersOnce(db, notifSvc)
		notificationReadCountDelta(t, participantToken, baseUnread, 1)
		notif := latestNotification(t, participantToken)
		if got := notificationType(notif); got != "event_reminder" {
			t.Fatalf("expected event_reminder notification, got %v", got)
		}
		if got := notificationRefID(notif); int(got) != int(eventID) {
			t.Fatalf("expected ref id %.0f, got %v", eventID, got)
		}
	})

	// ========================================================================
	// Scenario: mark_single_read & mark_all_read
	// ========================================================================
	t.Run("Mark Read Endpoints", func(t *testing.T) {
		baseUnread := getUnreadCount(t, tokenUserA)

		// Create two fresh followers so we deterministically produce +2 notifications
		s2 := newSuffix()
		tokenUserC, _ := registerAndLogin(t, s2, "User C Notif", "notifc_"+s2+"@test.com", "student", "FIK", "IF")
		tokenUserD, _ := registerAndLogin(t, s2, "User D Notif", "notifd_"+s2+"@test.com", "student", "FIK", "IF")

		status, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/users/%.0f/follow", idA), tokenUserC, nil)
		assertStatus(t, http.StatusCreated, status, res)

		status, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/users/%.0f/follow", idA), tokenUserD, nil)
		assertStatus(t, http.StatusCreated, status, res)

		// User A should have at least 2 additional unread notifications now
		count := getUnreadCount(t, tokenUserA)
		if count < baseUnread+2 {
			t.Errorf("Expected at least +2 unread, baseline %v now %v", baseUnread, count)
		}

		notifs := getNotifications(t, tokenUserA)
		if len(notifs) < 2 {
			t.Fatal("Need 2 notifications for test")
		}
		first := notifs[0].(map[string]any)
		firstIDAny := pickAny(first, "id", "ID")
		firstNotifID, ok := firstIDAny.(float64)
		if !ok || firstNotifID == 0 {
			t.Fatalf("Invalid notification ID: %v", firstIDAny)
		}

		// Try marking 1 as read
		status, res = jsonReq(t, http.MethodPatch, fmt.Sprintf("/api/notifications/%v/read", firstNotifID), tokenUserA, nil)
		assertStatus(t, http.StatusOK, status, res)

		// Count should go down by 1
		countMarkOne := getUnreadCount(t, tokenUserA)
		if countMarkOne >= count {
			t.Errorf("Expected count to decrease, got %v (was %v)", countMarkOne, count)
		}

		// Mark all as read
		status, res = jsonReq(t, http.MethodPatch, "/api/notifications/read-all", tokenUserA, nil)
		assertStatus(t, http.StatusOK, status, res)

		// Count should be 0
		countAllRead := getUnreadCount(t, tokenUserA)
		if countAllRead != 0 {
			t.Errorf("Expected 0 unread after read-all, got %v", countAllRead)
		}
	})

	// Note: We skip complex WS/Job/Mentor/Admin tests here to avoid 400 lines of setup,
	// but the underlying Notify/Throttled logics are perfectly unit tested by the patterns above.
}
