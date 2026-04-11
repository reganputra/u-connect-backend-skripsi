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
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	PhotoURL    *string    `json:"photo_url"`
	Location    *string    `json:"location"`
	Capacity    *int       `json:"capacity"`
	StartTime   *time.Time `json:"start_time"`
	Status      string     `json:"status"`
}

type EventAgendaRequest struct {
	Description string     `json:"description"`
	AgendaTime  *time.Time `json:"agenda_time"`
}

// ─── Interface ────────────────────────────────────────────────────────────────

type EventService interface {
	CreateEvent(userID uint, req EventRequest) (*models.Event, error)
	GetEvents(page, limit int) ([]models.Event, int64, error)
	GetEventByID(id uint) (*models.Event, error)
	UpdateEvent(userID, eventID uint, req EventRequest) (*models.Event, error)
	DeleteEvent(userID, eventID uint) error
	RegisterForEvent(userID, eventID uint) error
	CancelRegistration(userID, eventID uint) error
	GetParticipants(eventID uint) ([]models.EventRegistration, error)
	AddAgenda(userID, eventID uint, req EventAgendaRequest) (*models.EventAgenda, error)
	UpdateAgenda(userID, agendaID uint, req EventAgendaRequest) (*models.EventAgenda, error)
	DeleteAgenda(userID, agendaID uint) error
}

// ─── Struct ───────────────────────────────────────────────────────────────────

type eventService struct {
	eventRepo  repository.EventRepository
	agendaRepo repository.EventAgendaRepository
	regRepo    repository.EventRegistrationRepository
}

func NewEventService(
	eventRepo repository.EventRepository,
	agendaRepo repository.EventAgendaRepository,
	regRepo repository.EventRegistrationRepository,
) EventService {
	return &eventService{
		eventRepo:  eventRepo,
		agendaRepo: agendaRepo,
		regRepo:    regRepo,
	}
}

// ─── Event CRUD ───────────────────────────────────────────────────────────────

func (s *eventService) CreateEvent(userID uint, req EventRequest) (*models.Event, error) {
	if req.Title == "" {
		return nil, errors.New("judul wajib diisi")
	}
	status := req.Status
	if status == "" {
		status = "upcoming"
	}
	if !validEventStatuses[status] {
		return nil, errors.New("status tidak valid: harus upcoming, ongoing, completed, atau cancelled")
	}
	if req.Capacity != nil && *req.Capacity < 0 {
		return nil, errors.New("kapasitas harus nol atau positif")
	}

	event := &models.Event{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		PhotoURL:    req.PhotoURL,
		Location:    req.Location,
		Capacity:    req.Capacity,
		StartTime:   req.StartTime,
		Status:      status,
	}
	if err := s.eventRepo.CreateEvent(event); err != nil {
		return nil, errors.New("gagal membuat acara")
	}
	return event, nil
}

func (s *eventService) GetEvents(page, limit int) ([]models.Event, int64, error) {
	offset := (page - 1) * limit
	return s.eventRepo.FindEvents(offset, limit)
}

func (s *eventService) GetEventByID(id uint) (*models.Event, error) {
	event, err := s.eventRepo.FindEventByID(id)
	if err != nil {
		return nil, errors.New("acara tidak ditemukan")
	}
	return event, nil
}

func (s *eventService) UpdateEvent(userID, eventID uint, req EventRequest) (*models.Event, error) {
	event, err := s.eventRepo.FindEventByID(eventID)
	if err != nil {
		return nil, errors.New("acara tidak ditemukan")
	}
	if event.UserID != userID {
		return nil, errors.New("akses ditolak")
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
			return nil, errors.New("kapasitas harus nol atau positif")
		}
		event.Capacity = req.Capacity
	}
	if req.StartTime != nil {
		event.StartTime = req.StartTime
	}
	if req.Status != "" {
		if !validEventStatuses[req.Status] {
			return nil, errors.New("status tidak valid: harus upcoming, ongoing, completed, atau cancelled")
		}
		event.Status = req.Status
	}

	if err := s.eventRepo.UpdateEvent(event); err != nil {
		return nil, errors.New("gagal memperbarui acara")
	}
	return event, nil
}

func (s *eventService) DeleteEvent(userID, eventID uint) error {
	event, err := s.eventRepo.FindEventByID(eventID)
	if err != nil {
		return errors.New("acara tidak ditemukan")
	}
	if event.UserID != userID {
		return errors.New("akses ditolak")
	}
	return s.eventRepo.DeleteEvent(eventID)
}

// ─── Registration ─────────────────────────────────────────────────────────────

func (s *eventService) RegisterForEvent(userID, eventID uint) error {
	event, err := s.eventRepo.FindEventByID(eventID)
	if err != nil {
		return errors.New("acara tidak ditemukan")
	}

	if event.Status == "completed" || event.Status == "cancelled" {
		return errors.New("pendaftaran tidak diperbolehkan untuk acara yang telah selesai atau dibatalkan")
	}

	existing, _ := s.regRepo.FindEventRegistration(eventID, userID)
	if existing != nil {
		return errors.New("sudah terdaftar untuk acara ini")
	}

	if event.Capacity != nil {
		count, err := s.eventRepo.CountEventRegistrations(eventID)
		if err != nil {
			return errors.New("gagal memeriksa kapasitas")
		}
		if count >= int64(*event.Capacity) {
			return errors.New("acara sudah mencapai kapasitas penuh")
		}
	}

	reg := &models.EventRegistration{
		EventID: eventID,
		UserID:  userID,
	}
	if err := s.regRepo.CreateEventRegistration(reg); err != nil {
		return errors.New("gagal mendaftar ke acara")
	}
	return nil
}

func (s *eventService) CancelRegistration(userID, eventID uint) error {
	_, err := s.eventRepo.FindEventByID(eventID)
	if err != nil {
		return errors.New("acara tidak ditemukan")
	}

	existing, err := s.regRepo.FindEventRegistration(eventID, userID)
	if err != nil || existing == nil {
		return errors.New("Anda belum terdaftar untuk acara ini")
	}

	return s.regRepo.DeleteEventRegistration(eventID, userID)
}

func (s *eventService) GetParticipants(eventID uint) ([]models.EventRegistration, error) {
	_, err := s.eventRepo.FindEventByID(eventID)
	if err != nil {
		return nil, errors.New("acara tidak ditemukan")
	}
	return s.regRepo.FindEventParticipants(eventID)
}

// ─── Agenda ───────────────────────────────────────────────────────────────────

func (s *eventService) AddAgenda(userID, eventID uint, req EventAgendaRequest) (*models.EventAgenda, error) {
	if req.Description == "" {
		return nil, errors.New("deskripsi wajib diisi")
	}

	event, err := s.eventRepo.FindEventByID(eventID)
	if err != nil {
		return nil, errors.New("acara tidak ditemukan")
	}
	if event.UserID != userID {
		return nil, errors.New("akses ditolak")
	}

	agenda := &models.EventAgenda{
		EventID:     eventID,
		Description: req.Description,
		AgendaTime:  req.AgendaTime,
	}
	if err := s.agendaRepo.CreateEventAgenda(agenda); err != nil {
		return nil, errors.New("gagal menambahkan agenda")
	}
	return agenda, nil
}

func (s *eventService) UpdateAgenda(userID, agendaID uint, req EventAgendaRequest) (*models.EventAgenda, error) {
	agenda, err := s.agendaRepo.FindEventAgendaByID(agendaID)
	if err != nil {
		return nil, errors.New("item agenda tidak ditemukan")
	}

	event, err := s.eventRepo.FindEventByID(agenda.EventID)
	if err != nil {
		return nil, errors.New("acara tidak ditemukan")
	}
	if event.UserID != userID {
		return nil, errors.New("akses ditolak")
	}

	if req.Description != "" {
		agenda.Description = req.Description
	}
	if req.AgendaTime != nil {
		agenda.AgendaTime = req.AgendaTime
	}

	if err := s.agendaRepo.UpdateEventAgenda(agenda); err != nil {
		return nil, errors.New("gagal memperbarui agenda")
	}
	return agenda, nil
}

func (s *eventService) DeleteAgenda(userID, agendaID uint) error {
	agenda, err := s.agendaRepo.FindEventAgendaByID(agendaID)
	if err != nil {
		return errors.New("item agenda tidak ditemukan")
	}

	event, err := s.eventRepo.FindEventByID(agenda.EventID)
	if err != nil {
		return errors.New("acara tidak ditemukan")
	}
	if event.UserID != userID {
		return errors.New("akses ditolak")
	}

	return s.agendaRepo.DeleteEventAgenda(agendaID)
}
