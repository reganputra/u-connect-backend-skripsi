package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateGroupArticle(article *models.GroupArticle) error {
	return config.DB.Create(article).Error
}

func FindGroupArticles(groupID uint) ([]models.GroupArticle, error) {
	var articles []models.GroupArticle
	err := config.DB.
		Where("group_id = ?", groupID).
		Preload("User").
		Order("created_at desc").
		Find(&articles).Error
	return articles, err
}

func FindGroupArticleByID(id uint) (*models.GroupArticle, error) {
	var article models.GroupArticle
	err := config.DB.
		Preload("User").
		Preload("Reactions").
		First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func FindAllCommentsByArticleID(articleID uint) ([]models.GroupComment, error) {
	var comments []models.GroupComment
	err := config.DB.
		Where("article_id = ?", articleID).
		Preload("User").
		Preload("Reactions").
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func UpdateGroupArticle(article *models.GroupArticle) error {
	return config.DB.Save(article).Error
}

func DeleteGroupArticle(id uint) error {
	config.DB.Where("article_id = ?", id).Delete(&models.GroupReaction{})
	config.DB.Where("comment_id IN (SELECT id FROM group_comments WHERE article_id = ?)", id).Delete(&models.GroupReaction{})
	config.DB.Where("article_id = ?", id).Delete(&models.GroupComment{})
	return config.DB.Delete(&models.GroupArticle{}, id).Error
}
