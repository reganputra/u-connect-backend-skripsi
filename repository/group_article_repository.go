package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type GroupArticleRepository interface {
	CreateGroupArticle(article *models.GroupArticle) error
	FindGroupArticles(groupID uint) ([]models.GroupArticle, error)
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
		Order("created_at desc").
		Find(&articles).Error
	return articles, err
}

func (r *groupArticleRepository) FindGroupArticleByID(id uint) (*models.GroupArticle, error) {
	var article models.GroupArticle
	err := r.db.
		Preload("User").
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
