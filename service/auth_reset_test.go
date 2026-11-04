package service

import (
	"errors"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"strings"
	"testing"
	"time"
)

type resetErrorRepo struct {
	repository.UserRepository
	user          models.User
	fail          string
	wrotePassword bool
}

func (r *resetErrorRepo) FindUserByEmail(string) (*models.User, error) { u := r.user; return &u, nil }
func (r *resetErrorRepo) UnlockResetAttempts(uint) error {
	if r.fail == "unlock" {
		return errors.New("offline")
	}
	return nil
}
func (r *resetErrorRepo) IncrementResetAttempts(uint) error {
	if r.fail == "increment" {
		return errors.New("offline")
	}
	return nil
}
func (r *resetErrorRepo) LockResetWithTime(uint, time.Time) error {
	if r.fail == "lock" {
		return errors.New("offline")
	}
	return nil
}
func (r *resetErrorRepo) ResetPasswordAndClearAttempts(uint, string) error {
	r.wrotePassword = true
	return nil
}
func TestResetStateErrorsPropagate(t *testing.T) {
	for _, tc := range []struct {
		fail     string
		attempts int
		want     string
	}{
		{"unlock", 3, "gagal membuka kunci reset"},
		{"increment", 0, "gagal mencatat percobaan reset"},
		{"lock", 2, "gagal mengunci reset"},
	} {
		t.Run(tc.fail, func(t *testing.T) {
			old := time.Now().Add(-time.Hour)
			r := &resetErrorRepo{user: models.User{ResetAttempts: tc.attempts, ResetLockedAt: &old}, fail: tc.fail}
			svc := NewAuthService(r, nil)
			err := svc.ForgotPassword(ForgotPasswordRequest{Email: "auth@example.invalid", Faculty: "x", Major: "x", YearEnroll: 2020, NewPassword: "newpassword", ConfirmPassword: "newpassword"})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v want %s", err, tc.want)
			}
			if r.wrotePassword {
				t.Fatal("password written after state failure")
			}
		})
	}
}
