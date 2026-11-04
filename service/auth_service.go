package service

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/reganputra/skripsi-backend/constant"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/utils"
	"golang.org/x/crypto/bcrypt"
)

var validRoles = map[string]bool{
	constant.RoleAlumni:  true,
	constant.RoleStudent: true,
	constant.RolePartner: true,
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	// alumni & student
	Faculty    string `json:"faculty"`
	Major      string `json:"major"`
	YearEnroll int    `json:"year_enroll"`
	// partner
	CompanyName string `json:"company_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ForgotPasswordRequest struct {
	Email           string `json:"email"`
	Faculty         string `json:"faculty"`
	Major           string `json:"major"`
	YearEnroll      int    `json:"year_enroll"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type AuthUserPayload struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	IsActive    bool    `json:"is_active"`
	PictureURL  string  `json:"picture_url"`
	Faculty     *string `json:"faculty"`
	Major       *string `json:"major"`
	YearEnroll  *int    `json:"year_enroll"`
	CompanyName *string `json:"company_name"`
}

type LoginResponse struct {
	Token        string           `json:"token,omitempty"`
	RefreshToken string           `json:"-"`
	User         *AuthUserPayload `json:"user"`
}

type AuthClaims struct {
	Purpose      string `json:"token_type"`
	UserID       uint   `json:"user_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

// ─── Interface ────────────────────────────────────────────────────────────────

type AuthService interface {
	Register(req RegisterRequest) (*models.User, error)
	Login(req LoginRequest) (*LoginResponse, error)
	Refresh(refreshToken string) (*LoginResponse, error)
	Logout(userID uint) error
	Me(userID uint) (*AuthUserPayload, error)
	ChangePassword(userID uint, req ChangePasswordRequest) error
	ForgotPassword(req ForgotPasswordRequest) error
	UnlockUserReset(userID uint) error
}

// ─── Implementation ───────────────────────────────────────────────────────────

type authService struct {
	userRepo    repository.UserRepository
	profileRepo repository.ProfileRepository
}

func NewAuthService(userRepo repository.UserRepository, profileRepo repository.ProfileRepository) AuthService {
	return &authService{userRepo: userRepo, profileRepo: profileRepo}
}

func (s *authService) Register(req RegisterRequest) (*models.User, error) {
	// Validate common fields
	if req.Name == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		return nil, errors.New("nama, email, kata sandi, dan peran wajib diisi")
	}
	if !validRoles[req.Role] {
		return nil, errors.New("peran tidak valid: harus alumni, student, atau partner")
	}

	// Validate role-specific fields
	if req.Role == constant.RoleAlumni || req.Role == constant.RoleStudent {
		if req.Faculty == "" || req.Major == "" || req.YearEnroll == 0 {
			return nil, errors.New("fakultas, jurusan, dan tahun masuk wajib diisi untuk alumni dan mahasiswa")
		}
	}
	if req.Role == constant.RolePartner {
		req.CompanyName = strings.TrimSpace(req.CompanyName)
	}

	// Check email uniqueness
	existing, _ := s.userRepo.FindUserByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi kata sandi")
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
		Role:     req.Role,
	}

	// Assign role-specific fields
	if req.Role == constant.RoleAlumni || req.Role == constant.RoleStudent {
		user.Faculty = &req.Faculty
		user.Major = &req.Major
		user.YearEnroll = &req.YearEnroll
	}
	if req.Role == constant.RolePartner {
		if req.CompanyName != "" {
			user.CompanyName = &req.CompanyName
		}
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, errors.New("gagal membuat pengguna")
	}

	return user, nil
}

func (s *authService) Login(req LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email dan kata sandi wajib diisi")
	}

	user, err := s.userRepo.FindUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("email atau kata sandi tidak valid")
	}

	if !user.IsActive {
		return nil, errors.New("akun dinonaktifkan, hubungi admin")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau kata sandi tidak valid")
	}

	signed, refresh, err := s.issueTokens(user)
	if err != nil {
		return nil, err
	}

	authUser, err := s.buildAuthUser(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:        signed,
		RefreshToken: refresh,
		User:         authUser,
	}, nil
}

func (s *authService) issueTokens(user *models.User) (access, refresh string, err error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", "", errors.New("JWT_SECRET is not configured")
	}
	now := time.Now()

	accessClaims := AuthClaims{
		Purpose:      "access",
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(constant.JWTAccessExpiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessTok := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	access, err = accessTok.SignedString([]byte(secret))
	if err != nil {
		return "", "", errors.New("gagal membuat token")
	}

	refreshClaims := AuthClaims{
		Purpose:      "refresh",
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(constant.JWTRefreshExpiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refresh, err = refreshTok.SignedString([]byte(secret))
	if err != nil {
		return "", "", errors.New("gagal membuat refresh token")
	}
	return access, refresh, nil
}

func (s *authService) Refresh(refreshToken string) (*LoginResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token diperlukan")
	}
	claims, err := utils.ParseRefreshJWT(refreshToken)
	if err != nil {
		return nil, errors.New("refresh token tidak valid atau kadaluarsa")
	}
	id, version, err := utils.TokenIdentity(claims)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.FindUserByID(id)
	if err != nil || user == nil || !user.IsActive || user.TokenVersion != version || user.Role != claims["role"] {
		return nil, errors.New("sesi kadaluarsa, silakan login kembali")
	}
	// Sign the exact next generation before CAS. Never reread the winner's version.
	updated := *user
	updated.TokenVersion = version + 1
	signed, refresh, err := s.issueTokens(&updated)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.CompareAndSwapTokenVersion(id, version); err != nil {
		return nil, errors.New("sesi kadaluarsa, silakan login kembali")
	}
	authUser, err := s.buildAuthUser(&updated)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		Token:        signed,
		RefreshToken: refresh,
		User:         authUser,
	}, nil
}

func (s *authService) Logout(userID uint) error {
	return s.userRepo.BumpTokenVersion(userID)
}

func (s *authService) Me(userID uint) (*AuthUserPayload, error) {
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	return s.buildAuthUser(user)
}

func (s *authService) buildAuthUser(user *models.User) (*AuthUserPayload, error) {
	pictureURL := ""
	if s.profileRepo != nil {
		if profile, err := s.profileRepo.FindProfileByUserID(user.ID); err == nil {
			pictureURL = profile.ProfilePicture
		}
	}

	return &AuthUserPayload{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Role:        user.Role,
		IsActive:    user.IsActive,
		PictureURL:  pictureURL,
		Faculty:     user.Faculty,
		Major:       user.Major,
		YearEnroll:  user.YearEnroll,
		CompanyName: user.CompanyName,
	}, nil
}

func (s *authService) ChangePassword(userID uint, req ChangePasswordRequest) error {
	if req.OldPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		return errors.New("semua field wajib diisi")
	}
	if len(req.NewPassword) < 8 {
		return errors.New("password baru minimal 8 karakter")
	}
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("konfirmasi password tidak cocok")
	}

	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("password lama tidak sesuai")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal mengenkripsi password baru")
	}

	return s.userRepo.UpdateUserPassword(userID, string(hashed))
}

const maxResetAttempts = 3

func (s *authService) ForgotPassword(req ForgotPasswordRequest) error {
	if req.Email == "" || req.Faculty == "" || req.Major == "" || req.YearEnroll == 0 {
		return errors.New("email, fakultas, jurusan, dan tahun masuk wajib diisi")
	}
	if req.NewPassword == "" || req.ConfirmPassword == "" {
		return errors.New("password baru dan konfirmasi wajib diisi")
	}
	if len(req.NewPassword) < 8 {
		return errors.New("password baru minimal 8 karakter")
	}
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("konfirmasi password tidak cocok")
	}

	user, err := s.userRepo.FindUserByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		// Generic message — don't reveal whether email exists
		return errors.New("data yang dimasukkan tidak cocok")
	}

	// C-4: auto-expire lock after ForgotLockTTL
	if user.ResetAttempts >= maxResetAttempts {
		if user.ResetLockedAt == nil || time.Since(*user.ResetLockedAt) < time.Duration(constant.ForgotLockTTL)*time.Second {
			return errors.New("akun terkunci karena terlalu banyak percobaan reset, coba lagi dalam 30 menit atau hubungi admin")
		}
		// TTL expired — reset attempts and continue
		if err := s.userRepo.UnlockResetAttempts(user.ID); err != nil {
			return errors.New("gagal membuka kunci reset")
		}
		user.ResetAttempts = 0
		user.ResetLockedAt = nil
	}

	// Verify identity: faculty + major + year_enroll must match
	userFaculty := ""
	if user.Faculty != nil {
		userFaculty = strings.TrimSpace(*user.Faculty)
	}
	userMajor := ""
	if user.Major != nil {
		userMajor = strings.TrimSpace(*user.Major)
	}
	userYear := 0
	if user.YearEnroll != nil {
		userYear = *user.YearEnroll
	}

	if !strings.EqualFold(userFaculty, strings.TrimSpace(req.Faculty)) ||
		!strings.EqualFold(userMajor, strings.TrimSpace(req.Major)) ||
		userYear != req.YearEnroll {
		if err := s.userRepo.IncrementResetAttempts(user.ID); err != nil {
			return errors.New("gagal mencatat percobaan reset")
		}
		remaining := maxResetAttempts - user.ResetAttempts - 1
		if remaining <= 0 {
			now := time.Now()
			if err := s.userRepo.LockResetWithTime(user.ID, now); err != nil {
				return errors.New("gagal mengunci reset")
			}
			return errors.New("data tidak cocok — akun terkunci 30 menit, hubungi admin jika perlu")
		}
		return fmt.Errorf("data yang dimasukkan tidak cocok (%d percobaan tersisa)", remaining)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal mengenkripsi password")
	}

	return s.userRepo.ResetPasswordAndClearAttempts(user.ID, string(hashed))
}

func (s *authService) UnlockUserReset(userID uint) error {
	if _, err := s.userRepo.FindUserByID(userID); err != nil {
		return errors.New("pengguna tidak ditemukan")
	}
	return s.userRepo.UnlockResetAttempts(userID)
}
