package controllers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type MentorController struct {
	mentorSvc service.MentorService
}

func NewMentorController(mentorSvc service.MentorService) *MentorController {
	return &MentorController{mentorSvc: mentorSvc}
}

// ── Mentor Registration ────────────────────────────────────────────────────────

// RegisterAsMentor godoc — POST /api/mentor/register  (alumni only)
func (ctrl *MentorController) RegisterAsMentor(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	var req service.MentorRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	profile, err := ctrl.mentorSvc.RegisterAsMentor(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, profile)
}

// GetMyMentorProfile godoc — GET /api/mentor/profile  (alumni only)
func (ctrl *MentorController) GetMyMentorProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	profile, err := ctrl.mentorSvc.GetMyMentorProfile(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

// UpdateMentorProfile godoc — PUT /api/mentor/profile  (alumni only)
func (ctrl *MentorController) UpdateMentorProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	var req service.MentorRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	profile, err := ctrl.mentorSvc.UpdateMentorProfile(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

// UnregisterAsMentor godoc — DELETE /api/mentor/unregister  (alumni only)
func (ctrl *MentorController) UnregisterAsMentor(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	if err := ctrl.mentorSvc.UnregisterAsMentor(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil berhenti menjadi mentor"})
}

// ── Request Management (mentor-side) ──────────────────────────────────────────

// GetIncomingRequests godoc — GET /api/mentor/requests  (alumni only)
func (ctrl *MentorController) GetIncomingRequests(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	requests, err := ctrl.mentorSvc.GetIncomingRequests(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data permintaan")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, requests)
}

// ApproveRequest godoc — PATCH /api/mentor/requests/:id/approve  (alumni only)
func (ctrl *MentorController) ApproveRequest(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	reqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID permintaan tidak valid")
	}
	req, err := ctrl.mentorSvc.ApproveRequest(userID, uint(reqID))
	if err != nil {
		switch err.Error() {
		case "akses ditolak":
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		}
	}
	return utils.SuccessResponse(c, fiber.StatusOK, req)
}

// RejectRequest godoc — PATCH /api/mentor/requests/:id/reject  (alumni only)
func (ctrl *MentorController) RejectRequest(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	reqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID permintaan tidak valid")
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.BodyParser(&body)

	req, err := ctrl.mentorSvc.RejectRequest(userID, uint(reqID), body.Reason)
	if err != nil {
		switch err.Error() {
		case "akses ditolak":
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		}
	}
	return utils.SuccessResponse(c, fiber.StatusOK, req)
}

// GetMyMentees godoc — GET /api/mentor/mentees  (alumni only)
func (ctrl *MentorController) GetMyMentees(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	mentees, err := ctrl.mentorSvc.GetMyMentees(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data mentee")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, mentees)
}

// ── Session Management ────────────────────────────────────────────────────────

// CreateSession godoc — POST /api/mentor/sessions  (alumni only)
func (ctrl *MentorController) CreateSession(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	var body struct {
		StudentID   uint    `json:"student_id"`
		Topic       string  `json:"topic"`
		Notes       *string `json:"notes"`
		SessionDate *string `json:"session_date"` // ISO 8601 string → parsed below
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	req := service.SessionRequest{
		StudentID: body.StudentID,
		Topic:     body.Topic,
		Notes:     body.Notes,
	}
	if body.SessionDate != nil && *body.SessionDate != "" {
		t, err := time.Parse(time.RFC3339, *body.SessionDate)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "format session_date tidak valid (gunakan ISO 8601)")
		}
		req.SessionDate = &t
	}
	session, err := ctrl.mentorSvc.CreateSession(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, session)
}

// GetMentorSessions godoc — GET /api/mentor/sessions  (alumni only)
func (ctrl *MentorController) GetMentorSessions(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	sessions, err := ctrl.mentorSvc.GetMentorSessions(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil sesi mentoring")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, sessions)
}

// UpdateSession godoc — PATCH /api/mentor/sessions/:id  (alumni only)
func (ctrl *MentorController) UpdateSession(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	sessionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID sesi tidak valid")
	}
	var body struct {
		Topic       string  `json:"topic"`
		Notes       *string `json:"notes"`
		SessionDate *string `json:"session_date"`
		Status      string  `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	req := service.UpdateSessionRequest{
		Topic:  body.Topic,
		Notes:  body.Notes,
		Status: body.Status,
	}
	if body.SessionDate != nil && *body.SessionDate != "" {
		t, err := time.Parse(time.RFC3339, *body.SessionDate)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "format session_date tidak valid")
		}
		req.SessionDate = &t
	}
	session, err := ctrl.mentorSvc.UpdateSession(userID, uint(sessionID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, session)
}

// ── Student-side ──────────────────────────────────────────────────────────────

// GetMentors godoc — GET /api/mentors  (student only)
func (ctrl *MentorController) GetMentors(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	mentors, total, err := ctrl.mentorSvc.GetAvailableMentors(page, limit, search)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data mentor")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  mentors,
	})
}

// GetMentorDetail godoc — GET /api/mentors/:id  (student only)
func (ctrl *MentorController) GetMentorDetail(c *fiber.Ctx) error {
	mentorUserID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID mentor tidak valid")
	}
	profile, err := ctrl.mentorSvc.GetMentorDetail(uint(mentorUserID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

// RequestMentoring godoc — POST /api/mentors/:id/request  (student only)
func (ctrl *MentorController) RequestMentoring(c *fiber.Ctx) error {
	studentID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	mentorUserID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID mentor tidak valid")
	}
	var req service.MentoringRequestInput
	_ = c.BodyParser(&req) // message and similarity_score are optional

	mentorReq, err := ctrl.mentorSvc.RequestMentoring(studentID, uint(mentorUserID), req)
	if err != nil {
		if err.Error() == "anda sudah memiliki permintaan aktif atau sedang dibimbing oleh mentor ini" {
			return utils.ErrorResponse(c, fiber.StatusConflict, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, mentorReq)
}

// GetMyMentors godoc — GET /api/student/mentors  (student only)
func (ctrl *MentorController) GetMyMentors(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	mentors, err := ctrl.mentorSvc.GetMyMentors(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data mentor")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, mentors)
}

// GetSentRequests godoc — GET /api/student/requests  (student only)
func (ctrl *MentorController) GetSentRequests(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	requests, err := ctrl.mentorSvc.GetSentRequests(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data permintaan")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, requests)
}

// WithdrawRequest godoc — DELETE /api/student/requests/:id  (student only)
func (ctrl *MentorController) WithdrawRequest(c *fiber.Ctx) error {
	studentID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	reqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID permintaan tidak valid")
	}

	err = ctrl.mentorSvc.WithdrawRequest(studentID, uint(reqID))
	if err != nil {
		switch err.Error() {
		case "permintaan tidak ditemukan":
			return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
		case "akses ditolak":
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "permintaan mentoring berhasil ditarik"})
}

// CreateSessionAsStudent godoc — POST /api/student/sessions  (student only)
func (ctrl *MentorController) CreateSessionAsStudent(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	var body struct {
		MentorID    uint    `json:"mentor_id"`
		Topic       string  `json:"topic"`
		Notes       *string `json:"notes"`
		SessionDate *string `json:"session_date"`
	}
	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	req := service.StudentSessionRequest{
		MentorID: body.MentorID,
		Topic:    body.Topic,
		Notes:    body.Notes,
	}
	if body.SessionDate != nil && *body.SessionDate != "" {
		t, err := time.Parse(time.RFC3339, *body.SessionDate)
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "format session_date tidak valid (gunakan ISO 8601)")
		}
		req.SessionDate = &t
	}
	session, err := ctrl.mentorSvc.CreateSessionAsStudent(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, session)
}

// GetStudentSessions godoc — GET /api/student/sessions  (student only)
func (ctrl *MentorController) GetStudentSessions(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	sessions, err := ctrl.mentorSvc.GetStudentSessions(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil sesi mentoring")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, sessions)
}

// ── Recommendation ────────────────────────────────────────────────────────────

// GetRecommendations godoc — GET /api/mentors/recommend  (student only)
// Query param: ?q=custom+text  (omit for auto mode from profile)
// Query param: ?top=10         (number of results, default 10)
func (ctrl *MentorController) GetRecommendations(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	query := c.Query("q", "")
	topN, _ := strconv.Atoi(c.Query("top", "10"))

	results, err := ctrl.mentorSvc.GetRecommendations(userID, query, topN)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, results)
}
