package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type PostRepository interface {
	CreatePost(post *models.Post) error
	FindPosts(page, limit int) ([]models.Post, int64, error)
	FindPostByID(id uint) (*models.Post, error)
	FindAllCommentsByPostID(postID uint) ([]models.Comment, error)
	UpdatePost(post *models.Post) error
	DeletePost(id uint) error
	CreatePostImage(image *models.PostImage) error
	DeletePostImagesByPostID(postID uint) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) CreatePost(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *postRepository) FindPosts(page, limit int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * limit

	r.db.Model(&models.Post{}).Count(&total)

	err := r.db.
		Preload("User").
		Preload("Images").
		Preload("Reactions").
		Preload("Votes").
		Preload("Comments").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	return posts, total, err
}

func (r *postRepository) FindPostByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.
		Preload("User").
		Preload("Images").
		Preload("Reactions").
		Preload("Votes").
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) FindAllCommentsByPostID(postID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.
		Where("post_id = ?", postID).
		Preload("User").
		Preload("Reactions").
		Preload("Votes").
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func (r *postRepository) UpdatePost(post *models.Post) error {
	return r.db.Save(post).Error
}

func (r *postRepository) DeletePost(id uint) error {
	return r.db.Delete(&models.Post{}, id).Error
}

func (r *postRepository) CreatePostImage(image *models.PostImage) error {
	return r.db.Create(image).Error
}

func (r *postRepository) DeletePostImagesByPostID(postID uint) error {
	return r.db.Where("post_id = ?", postID).Delete(&models.PostImage{}).Error
}
