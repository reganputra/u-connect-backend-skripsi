package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type VoteRepository interface {
	FindVote(userID uint, postID *uint, commentID *uint) (*models.Vote, error)
	CreateVote(vote *models.Vote) error
	UpdateVote(vote *models.Vote) error
	DeleteVote(id uint) error
}

type voteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) VoteRepository {
	return &voteRepository{db: db}
}

func (r *voteRepository) FindVote(userID uint, postID *uint, commentID *uint) (*models.Vote, error) {
	var vote models.Vote
	query := r.db.Where("user_id = ?", userID)
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

func (r *voteRepository) CreateVote(vote *models.Vote) error {
	return r.db.Create(vote).Error
}

func (r *voteRepository) UpdateVote(vote *models.Vote) error {
	return r.db.Save(vote).Error
}

func (r *voteRepository) DeleteVote(id uint) error {
	return r.db.Delete(&models.Vote{}, id).Error
}
