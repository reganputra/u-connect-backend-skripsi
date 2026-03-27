package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type ReactionRepository interface {
	FindReaction(userID uint, postID *uint, commentID *uint) (*models.Reaction, error)
	CreateReaction(reaction *models.Reaction) error
	UpdateReaction(reaction *models.Reaction) error
	DeleteReaction(id uint) error
}

type reactionRepository struct {
	db *gorm.DB
}

func NewReactionRepository(db *gorm.DB) ReactionRepository {
	return &reactionRepository{db: db}
}

func (r *reactionRepository) FindReaction(userID uint, postID *uint, commentID *uint) (*models.Reaction, error) {
	var reaction models.Reaction
	query := r.db.Where("user_id = ?", userID)
	if postID != nil {
		query = query.Where("post_id = ?", *postID)
	} else if commentID != nil {
		query = query.Where("comment_id = ?", *commentID)
	}
	if err := query.First(&reaction).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *reactionRepository) CreateReaction(reaction *models.Reaction) error {
	return r.db.Create(reaction).Error
}

func (r *reactionRepository) UpdateReaction(reaction *models.Reaction) error {
	return r.db.Save(reaction).Error
}

func (r *reactionRepository) DeleteReaction(id uint) error {
	return r.db.Delete(&models.Reaction{}, id).Error
}
