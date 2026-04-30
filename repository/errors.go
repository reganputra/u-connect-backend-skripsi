package repository

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Sentinel errors returned by repository methods on constraint violations.
// Controllers should map these to 409 Conflict responses.
var (
	ErrAlreadyFollowing  = errors.New("already following")
	ErrAlreadyRegistered = errors.New("already registered")
)

// isDuplicateKeyError reports whether err is a PostgreSQL unique-constraint
// violation. Works for both GORM-wrapped errors and raw pg driver errors.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "uniqueviolation")
}
