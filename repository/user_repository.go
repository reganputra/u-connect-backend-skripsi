package repository

import (
	"time"

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
	LockResetWithTime(userID uint, t time.Time) error
	UnlockResetAttempts(userID uint) error
	ResetPasswordAndClearAttempts(userID uint, hashedPassword string) error
	// C-4: atomic token rotation
	BumpTokenVersion(userID uint) error
	CompareAndSwapTokenVersion(userID uint, expected int) error
	GetTokenVersion(userID uint) (int, error)
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
	return requireUserUpdate(r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password": hashedPassword, "token_version": gorm.Expr("token_version + 1"),
	}))
}

func (r *userRepository) IncrementResetAttempts(userID uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"reset_attempts":  gorm.Expr("reset_attempts + 1"),
			"reset_locked_at": gorm.Expr("CASE WHEN reset_attempts + 1 >= 3 THEN COALESCE(reset_locked_at, ?) ELSE reset_locked_at END", time.Now()),
		}).Error
}

func (r *userRepository) LockResetWithTime(userID uint, t time.Time) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{"reset_attempts": 3, "reset_locked_at": t}).Error
}

func (r *userRepository) UnlockResetAttempts(userID uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"reset_attempts": 0, "reset_locked_at": nil}).Error
}

func (r *userRepository) BumpTokenVersion(userID uint) error {
	return requireUserUpdate(r.db.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("token_version", gorm.Expr("token_version + 1")))
}

// CompareAndSwapTokenVersion is one SQL UPDATE: replay and concurrent refresh have one winner.
func (r *userRepository) CompareAndSwapTokenVersion(userID uint, expected int) error {
	return requireUserUpdate(r.db.Model(&models.User{}).
		Where("id = ? AND token_version = ? AND is_active = ?", userID, expected, true).
		UpdateColumn("token_version", gorm.Expr("token_version + 1")))
}

func requireUserUpdate(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *userRepository) GetTokenVersion(userID uint) (int, error) {
	var v int
	if err := r.db.Model(&models.User{}).Select("token_version").Where("id = ?", userID).Scan(&v).Error; err != nil {
		return 0, err
	}
	return v, nil
}

func (r *userRepository) ResetPasswordAndClearAttempts(userID uint, hashedPassword string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"password": hashedPassword, "reset_attempts": 0, "reset_locked_at": nil, "token_version": gorm.Expr("token_version + 1")}).Error
}
