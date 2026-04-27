package controllers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type EventController struct {
	eventSvc service.EventService
}

func NewEventController(eventSvc service.EventService) *EventController {
	return &EventController{eventSvc: eventSvc}
}

// ─── Event CRUD ───────────────────────────────────────────────────────────────

func (ctrl *EventController) CreateEvent(c *fiber.Ctx) error {
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
		Organizer:   parseOptionalString(c.FormValue("organizer")),
		Description: parseOptionalString(c.FormValue("description")),
		PhotoURL:    parseOptionalString(photoURL),
		Location:    parseOptionalString(c.FormValue("location")),
		Capacity:    capacity,
		StartTime:   parseEventTime(c.FormValue("start_time")),
		Status:      c.FormValue("status"),
	}

	event, err := ctrl.eventSvc.CreateEvent(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, event)
}

func (ctrl *EventController) GetEvents(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	events, total, err := ctrl.eventSvc.GetEvents(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data acara")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total":  total,
		"page":   page,
		"limit":  limit,
		"events": events,
	})
}

func (ctrl *EventController) GetMyOwnedEvents(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	events, total, err := ctrl.eventSvc.GetMyOwnedEvents(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data acara milik pengguna")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total":  total,
		"page":   page,
		"limit":  limit,
		"events": events,
	})
}

func (ctrl *EventController) GetMyRegisteredEvents(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	events, total, err := ctrl.eventSvc.GetMyRegisteredEvents(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data acara terdaftar pengguna")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total":  total,
		"page":   page,
		"limit":  limit,
		"events": events,
	})
}

func (ctrl *EventController) GetEventByID(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID acara tidak valid")
	}
	event, err := ctrl.eventSvc.GetEventByID(uint(eventID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, event)
}

func (ctrl *EventController) UpdateEvent(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID acara tidak valid")
	}

	photoURL, err := uploadFileIfPresent(c, "photo", "alumni-platform/events")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	req := service.EventRequest{
		Title:       c.FormValue("title"),
		Organizer:   parseOptionalString(c.FormValue("organizer")),
		Description: parseOptionalString(c.FormValue("description")),
		PhotoURL:    parseOptionalString(photoURL),
		Location:    parseOptionalString(c.FormValue("location")),
		Capacity:    parseOptionalInt(c.FormValue("capacity")),
		StartTime:   parseEventTime(c.FormValue("start_time")),
		Status:      c.FormValue("status"),
	}

	event, err := ctrl.eventSvc.UpdateEvent(userID, uint(eventID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, event)
}

func (ctrl *EventController) DeleteEvent(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID acara tidak valid")
	}
	if err := ctrl.eventSvc.DeleteEvent(userID, uint(eventID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "acara berhasil dihapus"})
}

// ─── Registration ─────────────────────────────────────────────────────────────

func (ctrl *EventController) RegisterForEvent(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID acara tidak valid")
	}
	if err := ctrl.eventSvc.RegisterForEvent(userID, uint(eventID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil mendaftar"})
}

func (ctrl *EventController) CancelRegistration(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID acara tidak valid")
	}
	if err := ctrl.eventSvc.CancelRegistration(userID, uint(eventID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "pendaftaran berhasil dibatalkan"})
}

func (ctrl *EventController) GetParticipants(c *fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID acara tidak valid")
	}
	participants, err := ctrl.eventSvc.GetParticipants(uint(eventID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, participants)
}

// ─── Agenda ───────────────────────────────────────────────────────────────────

func (ctrl *EventController) AddAgenda(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID acara tidak valid")
	}

	req := service.EventAgendaRequest{}
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	agenda, err := ctrl.eventSvc.AddAgenda(userID, uint(eventID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, agenda)
}

func (ctrl *EventController) UpdateAgenda(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	agendaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID agenda tidak valid")
	}

	req := service.EventAgendaRequest{}
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	agenda, err := ctrl.eventSvc.UpdateAgenda(userID, uint(agendaID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, agenda)
}

func (ctrl *EventController) DeleteAgenda(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	agendaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID agenda tidak valid")
	}
	if err := ctrl.eventSvc.DeleteAgenda(userID, uint(agendaID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "item agenda berhasil dihapus"})
}

// parseEventTime parses an RFC3339 time string into a *time.Time.
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
