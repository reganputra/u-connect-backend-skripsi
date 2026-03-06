package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

// FindVote finds a user's existing vote on a post or comment
func FindVote(userID uint, postID *uint, commentID *uint) (*models.Vote, error) {
	var vote models.Vote
	query := config.DB.Where("user_id = ?", userID)
	if postID != nil {
		query = query.Where("post_id = ?", *postID)
	} else if commentID != nil {
		query = query.Where("comment_id = ?", *commentID)
	}
	if err := query.First(&vote).Error; err != nil {
		return nil, err
	}
	return &vote, nil
}

func CreateVote(vote *models.Vote) error {
	return config.DB.Create(vote).Error
}

func UpdateVote(vote *models.Vote) error {
	return config.DB.Save(vote).Error
}

func DeleteVote(id uint) error {
	return config.DB.Delete(&models.Vote{}, id).Error
}
