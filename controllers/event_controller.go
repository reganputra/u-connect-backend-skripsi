package controllers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

// ─── Event CRUD ───────────────────────────────────────────────────────────────

func CreateEvent(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	photoURL, err := uploadFileIfPresent(c, "photo", "alumni-platform/events")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	capacity := parseOptionalInt(c.FormValue("capacity"))

	req := service.EventRequest{
		Title:       c.FormValue("title"),
		Description: parseOptionalString(c.FormValue("description")),
		PhotoURL:    parseOptionalString(photoURL),
		Location:    parseOptionalString(c.FormValue("location")),
		Capacity:    capacity,
		Status:      c.FormValue("status"),
	}

	event, err := service.CreateEvent(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, event)
}

func GetEvents(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	events, total, err := service.GetEvents(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to fetch events")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total":  total,
		"page":   page,
		"limit":  limit,
		"events": events,
	})
}

func GetEventByID(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid event id")
	}
	event, err := service.GetEventByID(uint(eventID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, event)
}

func UpdateEvent(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid event id")
	}

	photoURL, err := uploadFileIfPresent(c, "photo", "alumni-platform/events")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	req := service.EventRequest{
		Title:       c.FormValue("title"),
		Description: parseOptionalString(c.FormValue("description")),
		PhotoURL:    parseOptionalString(photoURL),
		Location:    parseOptionalString(c.FormValue("location")),
		Capacity:    parseOptionalInt(c.FormValue("capacity")),
		Status:      c.FormValue("status"),
	}

	event, err := service.UpdateEvent(userID, uint(eventID), req)
	if err != nil {
		if err.Error() == "access denied" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, event)
}

func DeleteEvent(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid event id")
	}
	if err := service.DeleteEvent(userID, uint(eventID)); err != nil {
		if err.Error() == "access denied" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "event deleted successfully"})
}

// ─── Registration ─────────────────────────────────────────────────────────────

func RegisterForEvent(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid event id")
	}
	if err := service.RegisterForEvent(userID, uint(eventID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "registered successfully"})
}

func CancelRegistration(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid event id")
	}
	if err := service.CancelRegistration(userID, uint(eventID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "registration cancelled"})
}

func GetParticipants(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid event id")
	}
	participants, err := service.GetParticipants(uint(eventID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, participants)
}

// ─── Agenda ───────────────────────────────────────────────────────────────────

func AddAgenda(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid event id")
	}

	req := service.EventAgendaRequest{}
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	agenda, err := service.AddAgenda(userID, uint(eventID), req)
	if err != nil {
		if err.Error() == "access denied" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, agenda)
}

func UpdateAgenda(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	agendaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid agenda id")
	}

	req := service.EventAgendaRequest{}
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}

	agenda, err := service.UpdateAgenda(userID, uint(agendaID), req)
	if err != nil {
		if err.Error() == "access denied" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, agenda)
}

func DeleteAgenda(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	agendaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid agenda id")
	}
	if err := service.DeleteAgenda(userID, uint(agendaID)); err != nil {
		if err.Error() == "access denied" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "agenda item deleted"})
}

// parseEventTime is a helper for parsing time strings in event tests.
// It lives here so it's available to the controller for future use.
func parseEventTime(val string) *time.Time {
	if val == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return nil
	}
	return &t
}
