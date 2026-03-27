package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type GroupReactionRepository interface {
	FindGroupReaction(userID uint, articleID *uint, commentID *uint) (*models.GroupReaction, error)
	CreateGroupReaction(reaction *models.GroupReaction) error
	UpdateGroupReaction(reaction *models.GroupReaction) error
	DeleteGroupReaction(id uint) error
}

type groupReactionRepository struct {
	db *gorm.DB
}

func NewGroupReactionRepository(db *gorm.DB) GroupReactionRepository {
	return &groupReactionRepository{db: db}
}

func (r *groupReactionRepository) FindGroupReaction(userID uint, articleID *uint, commentID *uint) (*models.GroupReaction, error) {
	var reaction models.GroupReaction
	query := r.db.Where("user_id = ?", userID)
	if articleID != nil {
		query = query.Where("article_id = ?", *articleID)
	} else if commentID != nil {
		query = query.Where("comment_id = ?", *commentID)
	}
	if err := query.First(&reaction).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *groupReactionRepository) CreateGroupReaction(reaction *models.GroupReaction) error {
	return r.db.Create(reaction).Error
}

func (r *groupReactionRepository) UpdateGroupReaction(reaction *models.GroupReaction) error {
	return r.db.Save(reaction).Error
}

func (r *groupReactionRepository) DeleteGroupReaction(id uint) error {
	return r.db.Delete(&models.GroupReaction{}, id).Error
}
