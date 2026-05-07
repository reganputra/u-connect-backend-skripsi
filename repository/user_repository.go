package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindUserByEmail(email string) (*models.User, error)
	FindUserByID(id uint) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUserCompanyName(userID uint, companyName string) error
	UpdateUserName(userID uint, name string) error
	UpdateUserPassword(userID uint, hashedPassword string) error
	IncrementResetAttempts(userID uint) error
	UnlockResetAttempts(userID uint) error
	ResetPasswordAndClearAttempts(userID uint, hashedPassword string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) UpdateUserCompanyName(userID uint, companyName string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("company_name", companyName).Error
}

func (r *userRepository) UpdateUserName(userID uint, name string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("name", name).Error
}

func (r *userRepository) UpdateUserPassword(userID uint, hashedPassword string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

func (r *userRepository) IncrementResetAttempts(userID uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		UpdateColumn("reset_attempts", gorm.Expr("reset_attempts + 1")).Error
}

func (r *userRepository) UnlockResetAttempts(userID uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"reset_attempts": 0, "reset_locked_at": nil}).Error
}

func (r *userRepository) ResetPasswordAndClearAttempts(userID uint, hashedPassword string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"password": hashedPassword, "reset_attempts": 0, "reset_locked_at": nil}).Error
}
