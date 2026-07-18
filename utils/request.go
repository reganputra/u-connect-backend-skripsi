package utils

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/dto"
)

// ParsePagination reads page/limit query params with caller-supplied defaults
// and clamps them to sane bounds. A non-positive page defaults to 1; a
// non-positive limit falls back to defaultLimit; an over-large limit is capped
// at MaxLimit from the constant package.
func ParsePagination(c *fiber.Ctx, defaultLimit int) (page, limit int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	if page <= 0 {
		page = 1
	}
	limit, _ = strconv.Atoi(c.Query("limit", strconv.Itoa(defaultLimit)))
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

// ParseIDParam parses a uint path/query param and returns a standardized error
// via ErrorResponse when it is missing or malformed. label is the resource name
// inserted into the message (e.g. "pengguna", "mentor").
func ParseIDParam(c *fiber.Ctx, param, label string) (uint, error) {
	raw := c.Params(param)
	if raw == "" {
		raw = c.Query(param, "")
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, &IDParamError{Label: label, Raw: raw}
	}
	return uint(id), nil
}

// IDParamError is returned by ParseIDParam; callers may render it directly.
type IDParamError struct {
	Label string
	Raw   string
}

func (e *IDParamError) Error() string {
	if e.Label != "" {
		return "ID " + e.Label + " tidak valid"
	}
	return dto.MsgInvalidUserID
}

// MustParseIDParam parses an ID param and writes the standardized 400 response
// itself, returning the parsed value. The boolean ok is false when the response
// was already sent (caller should return early).
func MustParseIDParam(c *fiber.Ctx, param, label string) (uint, bool) {
	id, err := ParseIDParam(c, param, label)
	if err != nil {
		_ = ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		return 0, false
	}
	return id, true
}

// ParseRFC3339 parses an RFC3339 timestamp, returning a standardized error
// when the value is empty or malformed.
func ParseRFC3339(value string) (*time.Time, error) {
	if value == "" {
		return nil, &TimeParseError{Value: value}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, &TimeParseError{Value: value}
	}
	return &t, nil
}

// TimeParseError is returned by ParseRFC3339.
type TimeParseError struct {
	Value string
}

func (e *TimeParseError) Error() string {
	return dto.MsgInvalidSessionDate
}

// ClampInt constrains v to [min, max], returning def when v is below min.
func ClampInt(v, min, max, def int) int {
	if v < min {
		return def
	}
	if v > max {
		return max
	}
	return v
}

// ParseEvaluationParams reads the shared top_n / student_ids query params used by
// the evaluation endpoints. topN is clamped to (0, max]; studentIDs is optional.
func ParseEvaluationParams(c *fiber.Ctx, max int) (topN int, studentIDs []uint) {
	topN, _ = strconv.Atoi(c.Query("top_n", "10"))
	if topN <= 0 || topN > max {
		topN = 10
	}
	if raw := c.Query("student_ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if id, err := strconv.ParseUint(part, 10, 64); err == nil && id > 0 {
				studentIDs = append(studentIDs, uint(id))
			}
		}
	}
	return topN, studentIDs
}
