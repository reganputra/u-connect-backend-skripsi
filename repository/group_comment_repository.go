package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type GroupCommentRepository interface {
	CreateGroupComment(comment *models.GroupComment) error
	FindGroupCommentByID(id uint) (*models.GroupComment, error)
	UpdateGroupComment(comment *models.GroupComment) error
	DeleteGroupComment(id uint) error
}

type groupCommentRepository struct {
	db *gorm.DB
}

func NewGroupCommentRepository(db *gorm.DB) GroupCommentRepository {
	return &groupCommentRepository{db: db}
}

func (r *groupCommentRepository) CreateGroupComment(comment *models.GroupComment) error {
	return r.db.Create(comment).Error
}

func (r *groupCommentRepository) FindGroupCommentByID(id uint) (*models.GroupComment, error) {
	var comment models.GroupComment
	if err := r.db.
		Preload("User").
		Preload("User.Profile").
		First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *groupCommentRepository) UpdateGroupComment(comment *models.GroupComment) error {
	return r.db.Save(comment).Error
}

func (r *groupCommentRepository) DeleteGroupComment(id uint) error {
	r.db.Where("comment_id = ?", id).Delete(&models.GroupReaction{})
	return r.db.Delete(&models.GroupComment{}, id).Error
}
