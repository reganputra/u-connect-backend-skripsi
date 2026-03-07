package service

import (
	"errors"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

var validEventStatuses = map[string]bool{
	"upcoming":  true,
	"ongoing":   true,
	"completed": true,
	"cancelled": true,
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type EventRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	PhotoURL    *string `json:"photo_url"` // set by controller after upload
	Location    *string `json:"location"`
	Capacity    *int    `json:"capacity"`
	Status      string  `json:"status"`
}

type EventAgendaRequest struct {
	Description string     `json:"description"`
	AgendaTime  *time.Time `json:"agenda_time"`
}

// ─── Event CRUD ───────────────────────────────────────────────────────────────

func CreateEvent(userID uint, req EventRequest) (*models.Event, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	status := req.Status
	if status == "" {
		status = "upcoming"
	}
	if !validEventStatuses[status] {
		return nil, errors.New("invalid status: must be upcoming, ongoing, completed, or cancelled")
	}
	if req.Capacity != nil && *req.Capacity < 0 {
		return nil, errors.New("capacity must be zero or positive")
	}

	event := &models.Event{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		PhotoURL:    req.PhotoURL,
		Location:    req.Location,
		Capacity:    req.Capacity,
		Status:      status,
	}
	if err := repository.CreateEvent(event); err != nil {
		return nil, errors.New("failed to create event")
	}
	return event, nil
}

func GetEvents(page, limit int) ([]models.Event, int64, error) {
	offset := (page - 1) * limit
	return repository.FindEvents(offset, limit)
}

func GetEventByID(id uint) (*models.Event, error) {
	event, err := repository.FindEventByID(id)
	if err != nil {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func UpdateEvent(userID, eventID uint, req EventRequest) (*models.Event, error) {
	event, err := repository.FindEventByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.UserID != userID {
		return nil, errors.New("access denied")
	}

	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != nil {
		event.Description = req.Description
	}
	if req.PhotoURL != nil {
		event.PhotoURL = req.PhotoURL
	}
	if req.Location != nil {
		event.Location = req.Location
	}
	if req.Capacity != nil {
		if *req.Capacity < 0 {
			return nil, errors.New("capacity must be zero or positive")
		}
		event.Capacity = req.Capacity
	}
	if req.Status != "" {
		if !validEventStatuses[req.Status] {
			return nil, errors.New("invalid status: must be upcoming, ongoing, completed, or cancelled")
		}
		event.Status = req.Status
	}

	if err := repository.UpdateEvent(event); err != nil {
		return nil, errors.New("failed to update event")
	}
	return event, nil
}

func DeleteEvent(userID, eventID uint) error {
	event, err := repository.FindEventByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.UserID != userID {
		return errors.New("access denied")
	}
	return repository.DeleteEvent(eventID)
}

// ─── Registration ─────────────────────────────────────────────────────────────

func RegisterForEvent(userID, eventID uint) error {
	event, err := repository.FindEventByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}

	if event.Status == "completed" || event.Status == "cancelled" {
		return errors.New("registration is not allowed for completed or cancelled events")
	}

	// Duplicate check
	existing, _ := repository.FindEventRegistration(eventID, userID)
	if existing != nil {
		return errors.New("already registered for this event")
	}

	// Capacity check
	if event.Capacity != nil {
		count, err := repository.CountEventRegistrations(eventID)
		if err != nil {
			return errors.New("failed to check capacity")
		}
		if count >= int64(*event.Capacity) {
			return errors.New("event is at full capacity")
		}
	}

	reg := &models.EventRegistration{
		EventID: eventID,
		UserID:  userID,
	}
	if err := repository.CreateEventRegistration(reg); err != nil {
		return errors.New("failed to register for event")
	}
	return nil
}

func CancelRegistration(userID, eventID uint) error {
	_, err := repository.FindEventByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}

	existing, err := repository.FindEventRegistration(eventID, userID)
	if err != nil || existing == nil {
		return errors.New("you are not registered for this event")
	}

	return repository.DeleteEventRegistration(eventID, userID)
}

func GetParticipants(eventID uint) ([]models.EventRegistration, error) {
	_, err := repository.FindEventByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	return repository.FindEventParticipants(eventID)
}

// ─── Agenda ───────────────────────────────────────────────────────────────────

func AddAgenda(userID, eventID uint, req EventAgendaRequest) (*models.EventAgenda, error) {
	if req.Description == "" {
		return nil, errors.New("description is required")
	}

	event, err := repository.FindEventByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.UserID != userID {
		return nil, errors.New("access denied")
	}

	agenda := &models.EventAgenda{
		EventID:     eventID,
		Description: req.Description,
		AgendaTime:  req.AgendaTime,
	}
	if err := repository.CreateEventAgenda(agenda); err != nil {
		return nil, errors.New("failed to add agenda")
	}
	return agenda, nil
}

func UpdateAgenda(userID, agendaID uint, req EventAgendaRequest) (*models.EventAgenda, error) {
	agenda, err := repository.FindEventAgendaByID(agendaID)
	if err != nil {
		return nil, errors.New("agenda item not found")
	}

	// Verify event ownership
	event, err := repository.FindEventByID(agenda.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.UserID != userID {
		return nil, errors.New("access denied")
	}

	if req.Description != "" {
		agenda.Description = req.Description
	}
	if req.AgendaTime != nil {
		agenda.AgendaTime = req.AgendaTime
	}

	if err := repository.UpdateEventAgenda(agenda); err != nil {
		return nil, errors.New("failed to update agenda")
	}
	return agenda, nil
}

func DeleteAgenda(userID, agendaID uint) error {
	agenda, err := repository.FindEventAgendaByID(agendaID)
	if err != nil {
		return errors.New("agenda item not found")
	}

	event, err := repository.FindEventByID(agenda.EventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.UserID != userID {
		return errors.New("access denied")
	}

	return repository.DeleteEventAgenda(agendaID)
}
