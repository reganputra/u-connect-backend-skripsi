package service

import (
	"errors"
	"os"
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

type AuthClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ─── Interface ────────────────────────────────────────────────────────────────

type AuthService interface {
	Register(req RegisterRequest) (*models.User, error)
	Login(req LoginRequest) (string, error)
}

// ─── Implementation ───────────────────────────────────────────────────────────

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
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
		if req.CompanyName == "" {
			return nil, errors.New("nama perusahaan wajib diisi untuk partner")
		}
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
		user.CompanyName = &req.CompanyName
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, errors.New("gagal membuat pengguna")
	}

	return user, nil
}

func (s *authService) Login(req LoginRequest) (string, error) {
	if req.Email == "" || req.Password == "" {
		return "", errors.New("email dan kata sandi wajib diisi")
	}

	user, err := s.userRepo.FindUserByEmail(req.Email)
	if err != nil {
		return "", errors.New("email atau kata sandi tidak valid")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("email atau kata sandi tidak valid")
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
		return "", errors.New("gagal membuat token")
	}

	return signed, nil
}
