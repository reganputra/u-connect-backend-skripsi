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

func Register(req RegisterRequest) (*models.User, error) {
	// Validate common fields
	if req.Name == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		return nil, errors.New("name, email, password, and role are required")
	}
	if !validRoles[req.Role] {
		return nil, errors.New("invalid role: must be alumni, student, or partner")
	}

	// Validate role-specific fields
	if req.Role == "alumni" || req.Role == "student" {
		if req.Faculty == "" || req.Major == "" || req.YearEnroll == 0 {
			return nil, errors.New("faculty, major, and year_enroll are required for alumni and student")
		}
	}
	if req.Role == "partner" {
		if req.CompanyName == "" {
			return nil, errors.New("company_name is required for partner")
		}
	}

	// Check email uniqueness
	existing, _ := repository.FindUserByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
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

	if err := repository.CreateUser(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func Login(req LoginRequest) (string, error) {
	if req.Email == "" || req.Password == "" {
		return "", errors.New("email and password are required")
	}

	user, err := repository.FindUserByEmail(req.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid email or password")
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
	if secret == "" {
		secret = "default_secret"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return signed, nil
}
