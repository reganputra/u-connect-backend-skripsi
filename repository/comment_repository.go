package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateComment(comment *models.Comment) error {
	return config.DB.Create(comment).Error
}

func FindCommentByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	if err := config.DB.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func UpdateComment(comment *models.Comment) error {
	return config.DB.Save(comment).Error
}

func DeleteComment(id uint) error {
	return config.DB.Delete(&models.Comment{}, id).Error
}
