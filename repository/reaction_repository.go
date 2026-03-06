package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

// FindReaction finds a user's existing reaction on a post or comment
func FindReaction(userID uint, postID *uint, commentID *uint) (*models.Reaction, error) {
	var reaction models.Reaction
	query := config.DB.Where("user_id = ?", userID)
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

func CreateReaction(reaction *models.Reaction) error {
	return config.DB.Create(reaction).Error
}

func UpdateReaction(reaction *models.Reaction) error {
	return config.DB.Save(reaction).Error
}

func DeleteReaction(id uint) error {
	return config.DB.Delete(&models.Reaction{}, id).Error
}
