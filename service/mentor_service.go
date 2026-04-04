package service

import (
	"errors"
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
	CreateSession(mentorUserID uint, req SessionRequest) (*models.MentoringSession, error)
	GetMentorSessions(mentorUserID uint) ([]models.MentoringSession, error)
	UpdateSession(mentorUserID, sessionID uint, req UpdateSessionRequest) (*models.MentoringSession, error)

	// ── Student-side ─────────────────────────────────────────────────────────
	GetAvailableMentors(page, limit int, search string) ([]models.UserProfile, int64, error)
	GetMentorDetail(mentorUserID uint) (*models.UserProfile, error)
	RequestMentoring(studentUserID, mentorUserID uint, req MentoringRequestInput) (*models.MentorRequest, error)
	GetMyMentors(studentUserID uint) ([]models.MentorRequest, error)
	GetSentRequests(studentUserID uint) ([]models.MentorRequest, error)
	GetStudentSessions(studentUserID uint) ([]models.MentoringSession, error)

	// ── Recommendation ────────────────────────────────────────────────────────
	// GetRecommendations returns TF-IDF ranked mentors.
	// If query is empty, it falls back to the student's profile skills+interests.
	GetRecommendations(studentUserID uint, query string, topN int) ([]RecommendResult, error)
}

// ─── Implementation ───────────────────────────────────────────────────────────

type mentorService struct {
	profileRepo    repository.ProfileRepository
	mentorRepo     repository.MentorRepository
	requestRepo    repository.MentorRequestRepository
	sessionRepo    repository.MentoringSessionRepository
	recommendSvc   RecommendationService
}

func NewMentorService(
	profileRepo repository.ProfileRepository,
	mentorRepo repository.MentorRepository,
	requestRepo repository.MentorRequestRepository,
	sessionRepo repository.MentoringSessionRepository,
	recommendSvc RecommendationService,
) MentorService {
	return &mentorService{
		profileRepo:  profileRepo,
		mentorRepo:   mentorRepo,
		requestRepo:  requestRepo,
		sessionRepo:  sessionRepo,
		recommendSvc: recommendSvc,
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

	// Check student hasn't exceeded 2-mentor limit
	count, _ := s.requestRepo.CountApprovedForStudent(req.StudentID)
	if count >= 2 {
		return nil, errors.New("mentee sudah memiliki 2 mentor aktif")
	}

	// Check mentor capacity
	mentorProfile, err := s.profileRepo.FindProfileByUserID(mentorUserID)
	if err != nil || mentorProfile.MentorQuota == nil {
		return nil, errors.New("profil mentor tidak ditemukan")
	}
	active, _ := s.mentorRepo.CountActiveMentees(mentorUserID)
	if active >= int64(*mentorProfile.MentorQuota) {
		return nil, errors.New("kapasitas mentor sudah penuh")
	}

	req.Status = "approved"
	if err := s.requestRepo.Update(req); err != nil {
		return nil, errors.New("gagal menyetujui permintaan")
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
	if err := s.requestRepo.Update(req); err != nil {
		return nil, errors.New("gagal menolak permintaan")
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

	validStatuses := map[string]bool{"scheduled": true, "completed": true, "cancelled": true}
	if req.Status != "" && !validStatuses[req.Status] {
		return nil, errors.New("status tidak valid")
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
		return nil, errors.New("gagal mengirim permintaan mentoring")
	}
	return mentorReq, nil
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
