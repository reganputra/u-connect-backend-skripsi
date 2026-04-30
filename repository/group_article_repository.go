package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type GroupArticleRepository interface {
	CreateGroupArticle(article *models.GroupArticle) error
	FindGroupArticles(groupID uint) ([]models.GroupArticle, error)
	// FindGroupArticlesPaginated returns paginated articles without comments;
	// use for group detail to avoid loading unbounded nested data.
	FindGroupArticlesPaginated(groupID uint, page, limit int) ([]models.GroupArticle, int64, error)
	FindGroupArticleByID(id uint) (*models.GroupArticle, error)
	FindAllCommentsByArticleID(articleID uint) ([]models.GroupComment, error)
	UpdateGroupArticle(article *models.GroupArticle) error
	DeleteGroupArticle(id uint) error
	// Image management
	FindArticleImages(articleID uint) ([]models.GroupArticleImage, error)
	DeleteArticleImages(articleID uint) error
	CreateArticleImage(img *models.GroupArticleImage) error
}

type groupArticleRepository struct {
	db *gorm.DB
}

func NewGroupArticleRepository(db *gorm.DB) GroupArticleRepository {
	return &groupArticleRepository{db: db}
}

func (r *groupArticleRepository) CreateGroupArticle(article *models.GroupArticle) error {
	return r.db.Create(article).Error
}

func (r *groupArticleRepository) FindGroupArticles(groupID uint) ([]models.GroupArticle, error) {
	var articles []models.GroupArticle
	err := r.db.
		Where("group_id = ?", groupID).
		Preload("User").
		Preload("User.Profile").
		Order("created_at desc").
		Find(&articles).Error
	return articles, err
}

// FindGroupArticlesPaginated returns articles for a group with pagination.
// It does NOT preload comments (use GetGroupArticleDetail for that),
// keeping this query bounded for use in group detail responses.
func (r *groupArticleRepository) FindGroupArticlesPaginated(groupID uint, page, limit int) ([]models.GroupArticle, int64, error) {
	var articles []models.GroupArticle
	var total int64

	base := r.db.Model(&models.GroupArticle{}).Where("group_id = ?", groupID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := r.db.
		Where("group_id = ?", groupID).
		Preload("User").
		Preload("User.Profile").
		Preload("Reactions").
		Order("created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&articles).Error
	return articles, total, err
}

func (r *groupArticleRepository) FindGroupArticleByID(id uint) (*models.GroupArticle, error) {
	var article models.GroupArticle
	err := r.db.
		Preload("User").
		Preload("User.Profile").
		Preload("Reactions").
		First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *groupArticleRepository) FindAllCommentsByArticleID(articleID uint) ([]models.GroupComment, error) {
	var comments []models.GroupComment
	err := r.db.
		Where("article_id = ?", articleID).
		Preload("User").
		Preload("User.Profile").
		Preload("Reactions").
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func (r *groupArticleRepository) UpdateGroupArticle(article *models.GroupArticle) error {
	return r.db.Save(article).Error
}

func (r *groupArticleRepository) DeleteGroupArticle(id uint) error {
	r.db.Where("article_id = ?", id).Delete(&models.GroupReaction{})
	r.db.Where("comment_id IN (SELECT id FROM group_comments WHERE article_id = ?)", id).Delete(&models.GroupReaction{})
	r.db.Where("article_id = ?", id).Delete(&models.GroupComment{})
	r.db.Where("article_id = ?", id).Delete(&models.GroupArticleImage{})
	return r.db.Delete(&models.GroupArticle{}, id).Error
}

// ─── Image management ─────────────────────────────────────────────────────────

func (r *groupArticleRepository) FindArticleImages(articleID uint) ([]models.GroupArticleImage, error) {
	var images []models.GroupArticleImage
	err := r.db.
		Where("article_id = ?", articleID).
		Order("created_at asc").
		Find(&images).Error
	return images, err
}

func (r *groupArticleRepository) DeleteArticleImages(articleID uint) error {
	return r.db.Where("article_id = ?", articleID).Delete(&models.GroupArticleImage{}).Error
}

func (r *groupArticleRepository) CreateArticleImage(img *models.GroupArticleImage) error {
	return r.db.Create(img).Error
}
