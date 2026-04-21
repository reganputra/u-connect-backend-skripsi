package service

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"golang.org/x/crypto/bcrypt"
)

var validRoles = map[string]bool{
	"alumni":  true,
	"student": true,
	"partner": true,
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
	Token string           `json:"token,omitempty"`
	User  *AuthUserPayload `json:"user"`
}

type AuthClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ─── Interface ────────────────────────────────────────────────────────────────

type AuthService interface {
	Register(req RegisterRequest) (*models.User, error)
	Login(req LoginRequest) (*LoginResponse, error)
	Me(userID uint) (*AuthUserPayload, error)
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
	if req.Role == "alumni" || req.Role == "student" {
		if req.Faculty == "" || req.Major == "" || req.YearEnroll == 0 {
			return nil, errors.New("fakultas, jurusan, dan tahun masuk wajib diisi untuk alumni dan mahasiswa")
		}
	}
	if req.Role == "partner" {
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
	if req.Role == "alumni" || req.Role == "student" {
		user.Faculty = &req.Faculty
		user.Major = &req.Major
		user.YearEnroll = &req.YearEnroll
	}
	if req.Role == "partner" {
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau kata sandi tidak valid")
	}

	claims := AuthClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	secret := os.Getenv("JWT_SECRET")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, errors.New("gagal membuat token")
	}

	authUser, err := s.buildAuthUser(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: signed,
		User:  authUser,
	}, nil
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
