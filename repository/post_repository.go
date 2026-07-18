package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// PostListRow is a lightweight projection for the feed list — counts only, no nested slices.
type PostListRow struct {
	models.Post
	CommentCount  int `gorm:"column:comment_count"`
	ReactionCount int `gorm:"column:reaction_count"`
	VoteScore     int `gorm:"column:vote_score"`
}

type PostRepository interface {
	CreatePost(post *models.Post) error
	FindPostsSummary(page, limit int) ([]PostListRow, int64, error)
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

// FindPostsSummary returns paginated posts with aggregated counts via SQL subqueries.
// Avoids N+1 caused by Preloading Comments/Reactions/Votes just to call len().
func (r *postRepository) FindPostsSummary(page, limit int) ([]PostListRow, int64, error) {
	var total int64
	if err := r.db.Model(&models.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []PostListRow
	offset := (page - 1) * limit
	err := r.db.Raw(`
		SELECT
			p.*,
			u.id         AS "User__id",
			u.name       AS "User__name",
			u.email      AS "User__email",
			u.role       AS "User__role",
			u.is_active  AS "User__is_active",
			up.profile_picture AS "User__picture_url",
			(SELECT COUNT(*) FROM comments  c WHERE c.post_id = p.id AND c.deleted_at IS NULL)           AS comment_count,
			(SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id AND r.deleted_at IS NULL)           AS reaction_count,
			(SELECT COALESCE(SUM(v.value),0) FROM votes v WHERE v.post_id = p.id AND v.deleted_at IS NULL) AS vote_score
		FROM posts p
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		LEFT JOIN user_profiles up ON up.user_id = u.id AND up.deleted_at IS NULL
		WHERE p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset).Scan(&rows).Error
	return rows, total, err
}

func (r *postRepository) FindPostByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.
		Preload("User").
		Preload("User.Profile").
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
		Preload("User.Profile").
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
