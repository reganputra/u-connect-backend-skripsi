package test

import (
	"fmt"
	"net/http"
	"testing"
)

// TestEvent covers event lifecycle, registration, capacity, and agendas.
// Run: go test -v -run TestEvent ./test/...
func TestEvent(t *testing.T) {
	sfx := newSuffix()
	ownerToken, _ := registerAndLogin(t, sfx, "Event Owner", "eventowner+"+sfx+"@test.com", "alumni", "Engineering", "CS")
	guestToken, _ := registerAndLogin(t, sfx, "Event Guest", "eventguest+"+sfx+"@test.com", "alumni", "Science", "Physics")
	partnerToken, _ := registerAndLoginPartner(t, "Event Partner "+sfx, "eventpartner+"+sfx+"@test.com", "PT Event Corp "+sfx)

	var eventID, agendaID, capacityEventID float64

	// ── Role Guard ─────────────────────────────────────────────────────────────

	t.Run("partner_cannot_create_event", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/events", partnerToken, map[string]string{
			"title": "should fail",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Partner blocked from creating event (403)")
	})

	t.Run("no_token_rejected", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/events", "", nil)
		assertStatus(t, 401, code, res)
		t.Log("✅ Unauthenticated request rejected (401)")
	})

	// ── Event CRUD ─────────────────────────────────────────────────────────────

	t.Run("create_event", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/events", ownerToken, map[string]string{
			"title":       "Alumni Seminar " + sfx,
			"description": "Annual alumni networking seminar",
			"location":    "Aula Besar, Kampus A",
			"capacity":    "50",
			"status":      "upcoming",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		eventID = d["ID"].(float64)
		t.Logf("✅ Event created (id: %.0f)", eventID)
	})

	t.Run("create_event_without_title_fails", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/events", ownerToken, map[string]string{
			"description": "missing title",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Event without title rejected (400)")
	})

	t.Run("create_event_invalid_status_fails", func(t *testing.T) {
		code, res := formReq(t, http.MethodPost, "/api/events", ownerToken, map[string]string{
			"title":  "Bad Status",
			"status": "random",
		})
		assertStatus(t, 400, code, res)
		t.Log("✅ Invalid status rejected (400)")
	})

	t.Run("get_events_list", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, "/api/events?page=1&limit=10", ownerToken, nil)
		assertStatus(t, 200, code, res)
		if res["data"] == nil {
			t.Fatal("missing data field in response")
		}
		t.Log("✅ Event list fetched")
	})

	t.Run("get_event_detail", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/events/%.0f", eventID), guestToken, nil)
		assertStatus(t, 200, code, res)
		d := dataMap(res)
		if d["Title"] == nil {
			t.Fatal("Title missing from event detail")
		}
		t.Log("✅ Event detail fetched")
	})

	t.Run("update_event", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/events/%.0f", eventID), ownerToken, map[string]string{
			"location": "Gedung Serbaguna Lt.2",
			"status":   "ongoing",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Event updated (owner)")
	})

	t.Run("non_owner_cannot_update_event", func(t *testing.T) {
		code, res := formReq(t, http.MethodPut, fmt.Sprintf("/api/events/%.0f", eventID), guestToken, map[string]string{
			"title": "hijacked",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner event update blocked (403)")
	})

	// ── Registration ───────────────────────────────────────────────────────────

	t.Run("register_for_event", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/register", eventID), guestToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Registered for event")
	})

	t.Run("cannot_register_twice", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/register", eventID), guestToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Duplicate registration rejected (400)")
	})

	t.Run("get_participants", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/events/%.0f/participants", eventID), ownerToken, nil)
		assertStatus(t, 200, code, res)
		participants, ok := res["data"].([]any)
		if !ok || len(participants) == 0 {
			t.Fatal("expected at least one participant")
		}
		t.Logf("✅ Participants fetched (%d)", len(participants))
	})

	// ── Capacity Enforcement ───────────────────────────────────────────────────

	t.Run("capacity_enforcement", func(t *testing.T) {
		// Create a small-capacity event (1 seat)
		sfx2 := newSuffix()
		code, res := formReq(t, http.MethodPost, "/api/events", ownerToken, map[string]string{
			"title":    "Full Event " + sfx2,
			"capacity": "1",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		capacityEventID = d["ID"].(float64)

		// First registration fills the seat
		code, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/register", capacityEventID), guestToken, nil)
		assertStatus(t, 200, code, res)

		// Second user tries — must be rejected
		extra1Token, _ := registerAndLogin(t, sfx2, "Extra User", "extra+"+sfx2+"@test.com", "alumni", "Law", "Law")
		code, res = jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/register", capacityEventID), extra1Token, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Capacity limit enforced (400)")
	})

	// ── Registration blocked on completed/cancelled ────────────────────────────

	t.Run("cannot_register_for_completed_event", func(t *testing.T) {
		// Create and immediately mark as completed
		sfx3 := newSuffix()
		_, createRes := formReq(t, http.MethodPost, "/api/events", ownerToken, map[string]string{
			"title":  "Done Event " + sfx3,
			"status": "completed",
		})
		doneID := dataMap(createRes)["ID"].(float64)
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/register", doneID), guestToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Registration blocked for completed event (400)")
	})

	t.Run("cannot_register_for_cancelled_event", func(t *testing.T) {
		sfx4 := newSuffix()
		_, createRes := formReq(t, http.MethodPost, "/api/events", ownerToken, map[string]string{
			"title":  "Cancelled Event " + sfx4,
			"status": "cancelled",
		})
		cancelledID := dataMap(createRes)["ID"].(float64)
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/register", cancelledID), guestToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Registration blocked for cancelled event (400)")
	})

	// ── Agenda ─────────────────────────────────────────────────────────────────

	t.Run("add_agenda", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/agenda", eventID), ownerToken, map[string]any{
			"description": "Opening Ceremony",
			"agenda_time": "2026-06-01T09:00:00Z",
		})
		assertStatus(t, 201, code, res)
		d := dataMap(res)
		agendaID = d["ID"].(float64)
		t.Logf("✅ Agenda item added (id: %.0f)", agendaID)
	})

	t.Run("non_owner_cannot_add_agenda", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPost, fmt.Sprintf("/api/events/%.0f/agenda", eventID), guestToken, map[string]any{
			"description": "should fail",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner agenda add blocked (403)")
	})

	t.Run("update_agenda", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, fmt.Sprintf("/api/agenda/%.0f", agendaID), ownerToken, map[string]any{
			"description": "Opening Ceremony & Welcome Speech",
		})
		assertStatus(t, 200, code, res)
		t.Log("✅ Agenda item updated")
	})

	t.Run("non_owner_cannot_update_agenda", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodPut, fmt.Sprintf("/api/agenda/%.0f", agendaID), guestToken, map[string]any{
			"description": "hijacked agenda",
		})
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner agenda update blocked (403)")
	})

	t.Run("delete_agenda", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/agenda/%.0f", agendaID), ownerToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Agenda item deleted")
	})

	// ── Cancel Registration ────────────────────────────────────────────────────

	t.Run("cancel_registration", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/events/%.0f/register", eventID), guestToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Registration cancelled")
	})

	t.Run("cancel_when_not_registered_fails", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/events/%.0f/register", eventID), guestToken, nil)
		assertStatus(t, 400, code, res)
		t.Log("✅ Cancel with no registration rejected (400)")
	})

	// ── Delete Event (cascade) ─────────────────────────────────────────────────

	t.Run("non_owner_cannot_delete_event", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/events/%.0f", eventID), guestToken, nil)
		assertStatus(t, 403, code, res)
		t.Log("✅ Non-owner event delete blocked (403)")
	})

	t.Run("delete_event", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodDelete, fmt.Sprintf("/api/events/%.0f", eventID), ownerToken, nil)
		assertStatus(t, 200, code, res)
		t.Log("✅ Event deleted (with cascade)")
	})

	t.Run("deleted_event_returns_404", func(t *testing.T) {
		code, res := jsonReq(t, http.MethodGet, fmt.Sprintf("/api/events/%.0f", eventID), ownerToken, nil)
		assertStatus(t, 404, code, res)
		t.Log("✅ Deleted event returns 404")
	})

	_ = partnerToken
	_ = capacityEventID
}
