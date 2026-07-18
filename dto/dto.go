// Package dto holds shared request/response contracts used by controllers and
// services. Centralizing these structs prevents the same JSON binding shape
// from being re-declared inline in multiple controllers.
package dto

// ─── Inline controller request DTOs (relocated from controllers) ────────────

// UpdateApplicationStatusRequest is the body for PATCH /api/jobs/:id/applications/:aid
type UpdateApplicationStatusRequest struct {
	Status string `json:"status"`
}

// ChangeCompanyAffiliationRequest is the body for company affiliation changes.
type ChangeCompanyAffiliationRequest struct {
	CompanyName string `json:"company_name"`
}

// KickMemberRequest is the body for kicking a group member.
type KickMemberRequest struct {
	Reason string `json:"reason"`
}

// RejectRequestRequest is the body for rejecting a mentor request.
type RejectRequestRequest struct {
	Reason string `json:"reason"`
}

// EndMentorshipRequest is the body for ending a mentorship (reason optional).
type EndMentorshipRequest struct {
	Reason *string `json:"reason"`
}

// MentorSessionRequest is the body for a mentor creating a session.
type MentorSessionRequest struct {
	StudentID   uint    `json:"student_id"`
	Topic       string  `json:"topic"`
	Notes       *string `json:"notes"`
	SessionDate *string `json:"session_date"`
}

// UpdateSessionRequestDTO is the body for updating a mentoring session.
type UpdateSessionRequestDTO struct {
	Topic       string  `json:"topic"`
	Notes       *string `json:"notes"`
	SessionDate *string `json:"session_date"`
	Status      string  `json:"status"`
}

// StudentSessionRequestDTO is the body for a student creating a session.
type StudentSessionRequestDTO struct {
	MentorID    uint    `json:"mentor_id"`
	Topic       string  `json:"topic"`
	Notes       *string `json:"notes"`
	SessionDate *string `json:"session_date"`
}

// SearchRecommendationsRequest is the body for mentor recommendation search.
type SearchRecommendationsRequest struct {
	Query string `json:"query"`
	Top   int    `json:"top"`
}

// ─── Pagination ──────────────────────────────────────────────────────────────

// PaginationMeta is included in responses with paginated sub-collections.
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewPaginationMeta builds a PaginationMeta from the raw counts.
func NewPaginationMeta(total int64, page, limit int) PaginationMeta {
	totalPages := int64(0)
	if limit > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}
	return PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: int(totalPages),
		HasNext:    int64(page) < totalPages,
		HasPrev:    page > 1,
	}
}

// ─── Shared message / error strings ─────────────────────────────────────────

const (
	// MsgInvalidRequest is the standard body-parser failure message.
	MsgInvalidRequest = "isi permintaan tidak valid"
	// MsgInvalidUserID is returned when an ID path/query param cannot be parsed.
	MsgInvalidUserID = "ID pengguna tidak valid"
	// MsgForbidden is the canonical "access denied" message.
	MsgForbidden = "akses ditolak"
	// MsgInvalidToken is returned when the JWT cannot be read.
	MsgInvalidToken = "token tidak valid"
	// MsgInvalidClaims is returned when JWT claims are malformed.
	MsgInvalidClaims = "klaim token tidak valid"
	// MsgUserIDClaimMissing is returned when user_id is missing from the token.
	MsgUserIDClaimMissing = "user_id tidak valid dalam token"
	// MsgRoleClaimMissing is returned when role is missing from the token.
	MsgRoleClaimMissing = "peran pengguna tidak ditemukan"
	// MsgInvalidSessionDate is returned when session_date is not RFC3339.
	MsgInvalidSessionDate = "format session_date tidak valid (gunakan ISO 8601)"
)
