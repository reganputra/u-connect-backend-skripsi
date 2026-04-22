package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

// ─── Request / Response types ─────────────────────────────────────────────────

type MentorRegisterRequest struct {
	MentorBio   string `json:"mentor_bio"`
	MentorQuota int    `json:"mentor_quota"` // allowed: 1, 2, 3, 5
}

type MentoringRequestInput struct {
	Message         *string  `json:"message"`
	SimilarityScore *float64 `json:"similarity_score"` // optional, recorded from recommendation
}

type SessionRequest struct {
	StudentID   uint       `json:"student_id"`
	Topic       string     `json:"topic"`
	Notes       *string    `json:"notes"`
	SessionDate *time.Time `json:"session_date"`
}

type StudentSessionRequest struct {
	MentorID    uint       `json:"mentor_id"`
	Topic       string     `json:"topic"`
	Notes       *string    `json:"notes"`
	SessionDate *time.Time `json:"session_date"`
}

type UpdateSessionRequest struct {
	Topic       string     `json:"topic"`
	Notes       *string    `json:"notes"`
	SessionDate *time.Time `json:"session_date"`
	Status      string     `json:"status"` // scheduled | completed | cancelled
}

// ─── Interface ────────────────────────────────────────────────────────────────

type MentorService interface {
	// ── Mentor-side ──────────────────────────────────────────────────────────
	RegisterAsMentor(userID uint, req MentorRegisterRequest) (*models.UserProfile, error)
	GetMyMentorProfile(userID uint) (*models.UserProfile, error)
	UpdateMentorProfile(userID uint, req MentorRegisterRequest) (*models.UserProfile, error)
	UnregisterAsMentor(userID uint) error
	GetIncomingRequests(mentorUserID uint) ([]models.MentorRequest, error)
	ApproveRequest(mentorUserID, requestID uint) (*models.MentorRequest, error)
	RejectRequest(mentorUserID, requestID uint, reason string) (*models.MentorRequest, error)
	GetMyMentees(mentorUserID uint) ([]models.MentorRequest, error)
	EndMentorship(mentorUserID, requestID uint, reason *string) (*models.MentorRequest, error)
	CreateSession(mentorUserID uint, req SessionRequest) (*models.MentoringSession, error)
	GetMentorSessions(mentorUserID uint) ([]models.MentoringSession, error)
	UpdateSession(mentorUserID, sessionID uint, req UpdateSessionRequest) (*models.MentoringSession, error)

	// ── Student-side ─────────────────────────────────────────────────────────
	GetAvailableMentors(page, limit int, search string) ([]models.UserProfile, int64, error)
	GetMentorDetail(mentorUserID uint) (*models.UserProfile, error)
	RequestMentoring(studentUserID, mentorUserID uint, req MentoringRequestInput) (*models.MentorRequest, error)
	WithdrawRequest(studentUserID, requestID uint) error
	GetMyMentors(studentUserID uint) ([]models.MentorRequest, error)
	GetSentRequests(studentUserID uint) ([]models.MentorRequest, error)
	CreateSessionAsStudent(studentUserID uint, req StudentSessionRequest) (*models.MentoringSession, error)
	GetStudentSessions(studentUserID uint) ([]models.MentoringSession, error)

	// ── Recommendation ────────────────────────────────────────────────────────
	// GetRecommendations returns TF-IDF ranked mentors.
	// If query is empty, it falls back to the student's profile skills+interests.
	GetRecommendations(studentUserID uint, query string, topN int) ([]RecommendResult, error)
}

// ─── Implementation ───────────────────────────────────────────────────────────

type mentorService struct {
	profileRepo  repository.ProfileRepository
	mentorRepo   repository.MentorRepository
	requestRepo  repository.MentorRequestRepository
	sessionRepo  repository.MentoringSessionRepository
	recommendSvc RecommendationService
	userRepo     repository.UserRepository
	notifSvc     NotificationService
}

func NewMentorService(
	profileRepo repository.ProfileRepository,
	mentorRepo repository.MentorRepository,
	requestRepo repository.MentorRequestRepository,
	sessionRepo repository.MentoringSessionRepository,
	recommendSvc RecommendationService,
	userRepo repository.UserRepository,
	notifSvc NotificationService,
) MentorService {
	return &mentorService{
		profileRepo:  profileRepo,
		mentorRepo:   mentorRepo,
		requestRepo:  requestRepo,
		sessionRepo:  sessionRepo,
		recommendSvc: recommendSvc,
		userRepo:     userRepo,
		notifSvc:     notifSvc,
	}
}

// allowedQuotas are the only valid mentor_quota values.
var allowedQuotas = map[int]bool{1: true, 2: true, 3: true, 5: true}

// ── Mentor-side ───────────────────────────────────────────────────────────────

func (s *mentorService) RegisterAsMentor(userID uint, req MentorRegisterRequest) (*models.UserProfile, error) {
	if req.MentorBio == "" {
		return nil, errors.New("mentor_bio wajib diisi")
	}
	if !allowedQuotas[req.MentorQuota] {
		return nil, errors.New("mentor_quota harus bernilai 1, 2, 3, atau 5")
	}

	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("profil tidak ditemukan — buat profil terlebih dahulu")
	}
	if profile.MentorQuota != nil {
		return nil, errors.New("anda sudah terdaftar sebagai mentor")
	}

	profile.MentorQuota = &req.MentorQuota
	profile.MentorDescription = &req.MentorBio

	if err := s.profileRepo.UpdateProfile(profile); err != nil {
		return nil, errors.New("gagal mendaftarkan sebagai mentor")
	}
	return profile, nil
}

func (s *mentorService) GetMyMentorProfile(userID uint) (*models.UserProfile, error) {
	profile, err := s.mentorRepo.FindMentorProfileByUserID(userID)
	if err != nil {
		return nil, errors.New("anda belum terdaftar sebagai mentor")
	}
	return profile, nil
}

func (s *mentorService) UpdateMentorProfile(userID uint, req MentorRegisterRequest) (*models.UserProfile, error) {
	if req.MentorQuota != 0 && !allowedQuotas[req.MentorQuota] {
		return nil, errors.New("mentor_quota harus bernilai 1, 2, 3, atau 5")
	}

	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil || profile.MentorQuota == nil {
		return nil, errors.New("anda belum terdaftar sebagai mentor")
	}

	// Check capacity: new quota must not be below current active mentees
	if req.MentorQuota != 0 {
		active, _ := s.mentorRepo.CountActiveMentees(userID)
		if int64(req.MentorQuota) < active {
			return nil, errors.New("kuota baru tidak boleh kurang dari jumlah mentee aktif")
		}
		profile.MentorQuota = &req.MentorQuota
	}
	if req.MentorBio != "" {
		profile.MentorDescription = &req.MentorBio
	}
	if err := s.profileRepo.UpdateProfile(profile); err != nil {
		return nil, errors.New("gagal memperbarui profil mentor")
	}
	return profile, nil
}

func (s *mentorService) UnregisterAsMentor(userID uint) error {
	active, err := s.mentorRepo.CountActiveMentees(userID)
	if err != nil {
		return errors.New("gagal memeriksa mentee aktif")
	}
	if active > 0 {
		return errors.New("tidak dapat berhenti menjadi mentor: masih memiliki mentee aktif")
	}

	profile, err := s.profileRepo.FindProfileByUserID(userID)
	if err != nil || profile.MentorQuota == nil {
		return errors.New("anda belum terdaftar sebagai mentor")
	}
	profile.MentorQuota = nil
	profile.MentorDescription = nil
	return s.profileRepo.UpdateProfile(profile)
}

func (s *mentorService) GetIncomingRequests(mentorUserID uint) ([]models.MentorRequest, error) {
	return s.requestRepo.FindByMentorID(mentorUserID)
}

func (s *mentorService) ApproveRequest(mentorUserID, requestID uint) (*models.MentorRequest, error) {
	req, err := s.requestRepo.ApprovePendingTransactional(mentorUserID, requestID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRequestNotFound):
			return nil, errors.New("permintaan tidak ditemukan")
		case errors.Is(err, repository.ErrRequestForbidden):
			return nil, errors.New("akses ditolak")
		case errors.Is(err, repository.ErrRequestAlreadyHandled):
			return nil, errors.New("permintaan sudah diproses")
		case errors.Is(err, repository.ErrStudentLimitReached):
			return nil, errors.New("mentee sudah memiliki 2 mentor aktif")
		case errors.Is(err, repository.ErrMentorProfileMissing):
			return nil, errors.New("profil mentor tidak ditemukan")
		case errors.Is(err, repository.ErrMentorCapacityFull):
			return nil, errors.New("kapasitas mentor sudah penuh")
		default:
			return nil, errors.New("gagal menyetujui permintaan")
		}
	}

	// Notify student
	if mentor, err := s.userRepo.FindUserByID(mentorUserID); err == nil {
		_ = s.notifSvc.Notify(
			req.StudentID,
			"mentor_request_approved",
			"Permintaan mentoring disetujui",
			fmt.Sprintf("Permintaan mentoringmu ke %s disetujui", mentor.Name),
			"mentor_request",
			req.ID,
		)
	}
	return req, nil
}

func (s *mentorService) RejectRequest(mentorUserID, requestID uint, reason string) (*models.MentorRequest, error) {
	req, err := s.requestRepo.FindByID(requestID)
	if err != nil {
		return nil, errors.New("permintaan tidak ditemukan")
	}
	if req.MentorID != mentorUserID {
		return nil, errors.New("akses ditolak")
	}
	if req.Status != "pending" {
		return nil, errors.New("permintaan sudah diproses")
	}

	req.Status = "rejected"
	req.RejectReason = &reason
	now := time.Now()
	req.RejectedAt = &now
	req.ApprovedAt = nil
	if err := s.requestRepo.Update(req); err != nil {
		return nil, errors.New("gagal menolak permintaan")
	}
	// Notify student
	if mentor, err := s.userRepo.FindUserByID(mentorUserID); err == nil {
		_ = s.notifSvc.Notify(
			req.StudentID,
			"mentor_request_rejected",
			"Permintaan mentoring ditolak",
			fmt.Sprintf("Permintaan mentoringmu ke %s ditolak", mentor.Name),
			"mentor_request",
			req.ID,
		)
	}
	return req, nil
}

func (s *mentorService) GetMyMentees(mentorUserID uint) ([]models.MentorRequest, error) {
	reqs, err := s.requestRepo.FindByMentorID(mentorUserID)
	if err != nil {
		return nil, err
	}
	// Filter to approved only
	var approved []models.MentorRequest
	for _, r := range reqs {
		if r.Status == "approved" {
			approved = append(approved, r)
		}
	}
	return approved, nil
}

func (s *mentorService) EndMentorship(mentorUserID, requestID uint, reason *string) (*models.MentorRequest, error) {
	req, cancelledSessions, err := s.requestRepo.EndMentorshipTransactional(mentorUserID, requestID, reason)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRequestNotFound):
			return nil, errors.New("permintaan tidak ditemukan")
		case errors.Is(err, repository.ErrRequestForbidden):
			return nil, errors.New("akses ditolak")
		case errors.Is(err, repository.ErrRequestAlreadyHandled):
			return nil, errors.New("mentorship sudah tidak aktif")
		default:
			return nil, errors.New("gagal mengakhiri mentorship")
		}
	}

	if mentor, err := s.userRepo.FindUserByID(mentorUserID); err == nil {
		message := fmt.Sprintf("Mentorship dengan %s telah diakhiri", mentor.Name)
		if cancelledSessions > 0 {
			message = fmt.Sprintf("Mentorship dengan %s telah diakhiri, %d sesi terjadwal dibatalkan", mentor.Name, cancelledSessions)
		}
		_ = s.notifSvc.Notify(
			req.StudentID,
			"mentor_relationship_ended",
			"Mentorship telah diakhiri",
			message,
			"mentor_request",
			req.ID,
		)
	}

	return req, nil
}

func (s *mentorService) CreateSession(mentorUserID uint, req SessionRequest) (*models.MentoringSession, error) {
	if strings.TrimSpace(req.Topic) == "" {
		return nil, errors.New("topik sesi wajib diisi")
	}
	if req.StudentID == 0 {
		return nil, errors.New("student_id wajib diisi")
	}

	// Validate an approved relationship exists
	approvedReq, err := s.sessionRepo.FindApprovedRequest(mentorUserID, req.StudentID)
	if err != nil {
		return nil, errors.New("tidak ada hubungan mentoring yang disetujui dengan student ini")
	}

	session := &models.MentoringSession{
		RequestID:   approvedReq.ID,
		MentorID:    mentorUserID,
		StudentID:   req.StudentID,
		Topic:       req.Topic,
		Notes:       req.Notes,
		SessionDate: req.SessionDate,
		Status:      "scheduled",
	}
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, errors.New("gagal membuat sesi mentoring")
	}
	// Notify student
	if mentor, err := s.userRepo.FindUserByID(mentorUserID); err == nil {
		_ = s.notifSvc.Notify(
			req.StudentID,
			"new_session",
			"Sesi mentoring dijadwalkan",
			fmt.Sprintf("%s menjadwalkan sesi: %s", mentor.Name, req.Topic),
			"mentor_request",
			session.ID,
		)
	}
	return session, nil
}

func (s *mentorService) CreateSessionAsStudent(studentUserID uint, req StudentSessionRequest) (*models.MentoringSession, error) {
	if strings.TrimSpace(req.Topic) == "" {
		return nil, errors.New("topik sesi wajib diisi")
	}
	if req.MentorID == 0 {
		return nil, errors.New("mentor_id wajib diisi")
	}

	// Validate an approved relationship exists.
	approvedReq, err := s.sessionRepo.FindApprovedRequest(req.MentorID, studentUserID)
	if err != nil {
		return nil, errors.New("tidak ada hubungan mentoring yang disetujui dengan mentor ini")
	}

	session := &models.MentoringSession{
		RequestID:   approvedReq.ID,
		MentorID:    req.MentorID,
		StudentID:   studentUserID,
		Topic:       req.Topic,
		Notes:       req.Notes,
		SessionDate: req.SessionDate,
		Status:      "scheduled",
	}
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, errors.New("gagal membuat sesi mentoring")
	}

	// Notify mentor.
	if student, err := s.userRepo.FindUserByID(studentUserID); err == nil {
		_ = s.notifSvc.Notify(
			req.MentorID,
			"new_session",
			"Sesi mentoring dijadwalkan",
			fmt.Sprintf("%s menjadwalkan sesi: %s", student.Name, req.Topic),
			"mentor_request",
			session.ID,
		)
	}

	return session, nil
}

func (s *mentorService) GetMentorSessions(mentorUserID uint) ([]models.MentoringSession, error) {
	return s.sessionRepo.FindByMentorID(mentorUserID)
}

func (s *mentorService) UpdateSession(mentorUserID, sessionID uint, req UpdateSessionRequest) (*models.MentoringSession, error) {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, errors.New("sesi tidak ditemukan")
	}
	if session.MentorID != mentorUserID {
		return nil, errors.New("akses ditolak")
	}
	if session.Status == "completed" || session.Status == "cancelled" {
		return nil, errors.New("status sesi sudah final")
	}

	validStatuses := map[string]bool{"scheduled": true, "completed": true, "cancelled": true}
	if req.Status != "" && !validStatuses[req.Status] {
		return nil, errors.New("status tidak valid")
	}
	if req.Status != "" {
		if session.Status == "scheduled" && req.Status == "completed" {
			if session.SessionDate != nil && session.SessionDate.After(time.Now()) {
				return nil, errors.New("sesi belum dapat diselesaikan sebelum waktunya")
			}
			now := time.Now()
			session.CompletedAt = &now
			session.CancelledAt = nil
		}
		if session.Status == "scheduled" && req.Status == "cancelled" {
			now := time.Now()
			session.CancelledAt = &now
		}
	}

	if req.Topic != "" {
		session.Topic = req.Topic
	}
	if req.Notes != nil {
		session.Notes = req.Notes
	}
	if req.SessionDate != nil {
		session.SessionDate = req.SessionDate
	}
	if req.Status != "" {
		session.Status = req.Status
	}

	if err := s.sessionRepo.Update(session); err != nil {
		return nil, errors.New("gagal memperbarui sesi")
	}
	return session, nil
}

// ── Student-side ──────────────────────────────────────────────────────────────

func (s *mentorService) GetAvailableMentors(page, limit int, search string) ([]models.UserProfile, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.mentorRepo.FindMentors(page, limit, search)
}

func (s *mentorService) GetMentorDetail(mentorUserID uint) (*models.UserProfile, error) {
	profile, err := s.mentorRepo.FindMentorProfileByUserID(mentorUserID)
	if err != nil {
		return nil, errors.New("mentor tidak ditemukan")
	}
	return profile, nil
}

func (s *mentorService) RequestMentoring(studentUserID, mentorUserID uint, req MentoringRequestInput) (*models.MentorRequest, error) {
	if studentUserID == mentorUserID {
		return nil, errors.New("tidak dapat me-request diri sendiri")
	}

	// Check mentor exists and has capacity
	mentorProfile, err := s.profileRepo.FindProfileByUserID(mentorUserID)
	if err != nil || mentorProfile.MentorQuota == nil {
		return nil, errors.New("mentor tidak ditemukan atau belum terdaftar sebagai mentor")
	}
	active, _ := s.mentorRepo.CountActiveMentees(mentorUserID)
	if active >= int64(*mentorProfile.MentorQuota) {
		return nil, errors.New("kapasitas mentor sudah penuh")
	}

	// Duplicate pending/approved check
	_, err = s.requestRepo.FindActiveDuplicate(studentUserID, mentorUserID)
	if err == nil {
		return nil, errors.New("anda sudah memiliki permintaan aktif atau sedang dibimbing oleh mentor ini")
	}

	// Student 2-mentor limit check
	count, _ := s.requestRepo.CountApprovedForStudent(studentUserID)
	if count >= 2 {
		return nil, errors.New("anda sudah memiliki 2 mentor aktif")
	}

	mentorReq := &models.MentorRequest{
		MentorID:        mentorUserID,
		StudentID:       studentUserID,
		Message:         req.Message,
		Status:          "pending",
		SimilarityScore: req.SimilarityScore,
	}
	if err := s.requestRepo.Create(mentorReq); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "idx_active_request_pair") || strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return nil, errors.New("anda sudah memiliki permintaan aktif atau sedang dibimbing oleh mentor ini")
		}
		return nil, errors.New("gagal mengirim permintaan mentoring")
	}
	// Notify mentor
	if student, err := s.userRepo.FindUserByID(studentUserID); err == nil {
		_ = s.notifSvc.Notify(
			mentorUserID,
			"mentor_request_received",
			"Permintaan mentoring baru",
			fmt.Sprintf("%s mengirim permintaan mentoring", student.Name),
			"mentor_request",
			mentorReq.ID,
		)
	}
	return mentorReq, nil
}

func (s *mentorService) WithdrawRequest(studentUserID, requestID uint) error {
	req, err := s.requestRepo.FindByID(requestID)
	if err != nil {
		return errors.New("permintaan tidak ditemukan")
	}
	if req.StudentID != studentUserID {
		return errors.New("akses ditolak")
	}
	if req.Status != "pending" {
		return errors.New("hanya permintaan pending yang bisa ditarik")
	}
	now := time.Now()
	req.Status = "withdrawn"
	req.WithdrawnAt = &now
	if err := s.requestRepo.Update(req); err != nil {
		return errors.New("gagal menarik permintaan")
	}
	return nil
}

func (s *mentorService) GetMyMentors(studentUserID uint) ([]models.MentorRequest, error) {
	reqs, err := s.requestRepo.FindByStudentID(studentUserID)
	if err != nil {
		return nil, err
	}
	var approved []models.MentorRequest
	for _, r := range reqs {
		if r.Status == "approved" {
			approved = append(approved, r)
		}
	}
	return approved, nil
}

func (s *mentorService) GetSentRequests(studentUserID uint) ([]models.MentorRequest, error) {
	return s.requestRepo.FindByStudentID(studentUserID)
}

func (s *mentorService) GetStudentSessions(studentUserID uint) ([]models.MentoringSession, error) {
	return s.sessionRepo.FindByStudentID(studentUserID)
}

// ── Recommendation ────────────────────────────────────────────────────────────

func (s *mentorService) GetRecommendations(studentUserID uint, query string, topN int) ([]RecommendResult, error) {
	if strings.TrimSpace(query) == "" {
		// Auto mode: build student text from their profile skills + interests
		profile, err := s.profileRepo.FindProfileByUserID(studentUserID)
		if err != nil {
			return nil, errors.New("profil tidak ditemukan — buat profil terlebih dahulu")
		}
		parts := []string{}
		if profile.Skills != nil {
			parts = append(parts, *profile.Skills)
		}
		if profile.Interests != nil {
			parts = append(parts, *profile.Interests)
		}
		query = strings.Join(parts, " ")
	}

	if topN <= 0 {
		topN = 10
	}
	return s.recommendSvc.RecommendMentors(query, topN)
}
