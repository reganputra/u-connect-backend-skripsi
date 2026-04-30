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
	Organizer   *string    `json:"organizer"`
	Description *string    `json:"description"`
	PhotoURL    *string    `json:"photo_url"`
	Location    *string    `json:"location"`
	Capacity    *int       `json:"capacity"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
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
	GetMyOwnedEvents(userID uint, page, limit int) ([]models.Event, int64, error)
	GetMyRegisteredEvents(userID uint, page, limit int) ([]models.Event, int64, error)
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

func applyEventUserPicture(user *models.User) {
	if user == nil {
		return
	}
	if user.Profile != nil {
		if user.Profile.ProfilePicture != "" {
			picture := user.Profile.ProfilePicture
			user.PictureURL = &picture
		} else {
			user.PictureURL = nil
		}
		user.Profile = nil
	}
}

func hydrateEventUsers(event *models.Event) {
	if event == nil {
		return
	}
	applyEventUserPicture(&event.User)
	for i := range event.Registrations {
		applyEventUserPicture(&event.Registrations[i].User)
	}
}

func hydrateParticipantUsers(regs []models.EventRegistration) {
	for i := range regs {
		applyEventUserPicture(&regs[i].User)
	}
}

func calculateEventSeatLeft(capacity *int, registeredCount int64) *int {
	if capacity == nil || *capacity <= 0 {
		return nil
	}

	left := *capacity - int(registeredCount)
	if left < 0 {
		left = 0
	}

	return &left
}

func (s *eventService) attachEventRegistrationStats(event *models.Event) error {
	count, err := s.eventRepo.CountEventRegistrations(event.ID)
	if err != nil {
		return err
	}

	event.AttendantCount = count
	event.SeatLeft = calculateEventSeatLeft(event.Capacity, count)
	return nil
}

func (s *eventService) attachEventsRegistrationStats(events []models.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Collect event IDs and fetch all counts in one query (avoids N+1).
	eventIDs := make([]uint, 0, len(events))
	for i := range events {
		eventIDs = append(eventIDs, events[i].ID)
	}
	counts, err := s.eventRepo.CountEventRegistrationsBatch(eventIDs)
	if err != nil {
		return err
	}

	for i := range events {
		count := counts[events[i].ID] // zero if not in map (no registrations)
		events[i].AttendantCount = count
		events[i].SeatLeft = calculateEventSeatLeft(events[i].Capacity, count)
	}

	return nil
}


func (s *eventService) syncPastEventStatuses() error {
	_, err := s.eventRepo.AutoCompletePastEvents(time.Now())
	return err
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
		Organizer:   req.Organizer,
		Description: req.Description,
		PhotoURL:    req.PhotoURL,
		Location:    req.Location,
		Capacity:    req.Capacity,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      status,
	}
	if err := s.eventRepo.CreateEvent(event); err != nil {
		return nil, errors.New("gagal membuat acara")
	}

	event.AttendantCount = 0
	event.SeatLeft = calculateEventSeatLeft(event.Capacity, 0)
	return event, nil
}

func (s *eventService) GetEvents(page, limit int) ([]models.Event, int64, error) {
	if err := s.syncPastEventStatuses(); err != nil {
		return nil, 0, errors.New("gagal memperbarui status acara")
	}
	offset := (page - 1) * limit
	events, total, err := s.eventRepo.FindEvents(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	if err := s.attachEventsRegistrationStats(events); err != nil {
		return nil, 0, errors.New("gagal menghitung peserta acara")
	}

	return events, total, nil
}

func (s *eventService) GetMyOwnedEvents(userID uint, page, limit int) ([]models.Event, int64, error) {
	if err := s.syncPastEventStatuses(); err != nil {
		return nil, 0, errors.New("gagal memperbarui status acara")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	events, total, err := s.eventRepo.FindEventsByOwner(userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	if err := s.attachEventsRegistrationStats(events); err != nil {
		return nil, 0, errors.New("gagal menghitung peserta acara")
	}

	return events, total, nil
}

func (s *eventService) GetMyRegisteredEvents(userID uint, page, limit int) ([]models.Event, int64, error) {
	if err := s.syncPastEventStatuses(); err != nil {
		return nil, 0, errors.New("gagal memperbarui status acara")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	events, total, err := s.regRepo.FindRegisteredEventsByUser(userID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	if err := s.attachEventsRegistrationStats(events); err != nil {
		return nil, 0, errors.New("gagal menghitung peserta acara")
	}

	return events, total, nil
}

func (s *eventService) GetEventByID(id uint) (*models.Event, error) {
	if err := s.syncPastEventStatuses(); err != nil {
		return nil, errors.New("gagal memperbarui status acara")
	}
	event, err := s.eventRepo.FindEventByID(id)
	if err != nil {
		return nil, errors.New("acara tidak ditemukan")
	}

	if err := s.attachEventRegistrationStats(event); err != nil {
		return nil, errors.New("gagal menghitung peserta acara")
	}
	hydrateEventUsers(event)

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
	if req.Organizer != nil {
		event.Organizer = req.Organizer
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
	if req.EndTime != nil {
		event.EndTime = req.EndTime
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

	if err := s.attachEventRegistrationStats(event); err != nil {
		return nil, errors.New("gagal menghitung peserta acara")
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
	if err := s.syncPastEventStatuses(); err != nil {
		return errors.New("gagal memperbarui status acara")
	}
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

	if event.Capacity != nil && *event.Capacity > 0 {
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
	regs, err := s.regRepo.FindEventParticipants(eventID)
	if err != nil {
		return nil, err
	}
	hydrateParticipantUsers(regs)
	return regs, nil
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
