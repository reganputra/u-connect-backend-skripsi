package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreatePost(post *models.Post) error {
	return config.DB.Create(post).Error
}

func FindPosts(page, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * limit

	config.DB.Model(&models.Post{}).Count(&total)

	err := config.DB.
		Preload("User").
		Preload("Reactions").
		Preload("Votes").
		Preload("Comments").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, total, err
}

func FindPostByID(id uint) (*models.Post, error) {
	var post models.Post
	err := config.DB.
		Preload("User").
		Preload("Reactions").
		Preload("Votes").
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// FindAllCommentsByPostID loads every comment for a post (all depths) in one query,
// including User, Reactions, and Votes so they can be tree-built in the service layer.
func FindAllCommentsByPostID(postID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := config.DB.
		Where("post_id = ?", postID).
		Preload("User").
		Preload("Reactions").
		Preload("Votes").
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func UpdatePost(post *models.Post) error {
	return config.DB.Save(post).Error
}

func DeletePost(id uint) error {
	return config.DB.Delete(&models.Post{}, id).Error
}
