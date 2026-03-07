package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func FindGroupReaction(userID uint, articleID *uint, commentID *uint) (*models.GroupReaction, error) {
	var reaction models.GroupReaction
	query := config.DB.Where("user_id = ?", userID)
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

func CreateGroupReaction(reaction *models.GroupReaction) error {
	return config.DB.Create(reaction).Error
}

func UpdateGroupReaction(reaction *models.GroupReaction) error {
	return config.DB.Save(reaction).Error
}

func DeleteGroupReaction(id uint) error {
	return config.DB.Delete(&models.GroupReaction{}, id).Error
}
