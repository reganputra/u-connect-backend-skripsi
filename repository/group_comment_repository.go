package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateGroupComment(comment *models.GroupComment) error {
	return config.DB.Create(comment).Error
}

func FindGroupCommentByID(id uint) (*models.GroupComment, error) {
	var comment models.GroupComment
	if err := config.DB.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func UpdateGroupComment(comment *models.GroupComment) error {
	return config.DB.Save(comment).Error
}

func DeleteGroupComment(id uint) error {
	config.DB.Where("comment_id = ?", id).Delete(&models.GroupReaction{})
	return config.DB.Delete(&models.GroupComment{}, id).Error
}
