package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type FollowRepository interface {
	// Follow creates a follow record (followerID → followingID).
	Follow(followerID, followingID uint) error
	// Unfollow removes the follow record.
	Unfollow(followerID, followingID uint) error
	// IsFollowing returns true if followerID follows followingID.
	IsFollowing(followerID, followingID uint) (bool, error)
	// AreConnected returns true if a follow relationship exists in either direction.
	// Used for symmetric messaging permission check.
	AreConnected(userA, userB uint) (bool, error)
	// GetFollowers returns all users who follow userID.
	GetFollowers(userID uint) ([]models.User, error)
	// GetFollowing returns all users that userID follows.
	GetFollowing(userID uint) ([]models.User, error)
}

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) Follow(followerID, followingID uint) error {
	follow := models.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	if err := r.db.Create(&follow).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrAlreadyFollowing
		}
		return err
	}
	return nil
}

func (r *followRepository) Unfollow(followerID, followingID uint) error {
	return r.db.
		Unscoped().
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.Follow{}).Error
}

func (r *followRepository) IsFollowing(followerID, followingID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

func (r *followRepository) AreConnected(userA, userB uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).
		Where(
			"(follower_id = ? AND following_id = ?) OR (follower_id = ? AND following_id = ?)",
			userA, userB, userB, userA,
		).
		Count(&count).Error
	return count > 0, err
}

func (r *followRepository) GetFollowers(userID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Preload("Profile").
		Joins("JOIN follows f ON f.follower_id = users.id AND f.deleted_at IS NULL").
		Where("f.following_id = ? AND users.deleted_at IS NULL", userID).
		Find(&users).Error

	if err == nil {
		for i := range users {
			if users[i].Profile != nil && users[i].Profile.ProfilePicture != "" {
				users[i].PictureURL = &users[i].Profile.ProfilePicture
			}
		}
	}
	return users, err
}

func (r *followRepository) GetFollowing(userID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Preload("Profile").
		Joins("JOIN follows f ON f.following_id = users.id AND f.deleted_at IS NULL").
		Where("f.follower_id = ? AND users.deleted_at IS NULL", userID).
		Find(&users).Error

	if err == nil {
		for i := range users {
			if users[i].Profile != nil && users[i].Profile.ProfilePicture != "" {
				users[i].PictureURL = &users[i].Profile.ProfilePicture
			}
		}
	}
	return users, err
}
