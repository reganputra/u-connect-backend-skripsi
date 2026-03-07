package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreateGroup(group *models.Group) error {
	return config.DB.Create(group).Error
}

func FindGroups() ([]models.Group, error) {
	var groups []models.Group
	err := config.DB.Preload("Owner").Order("created_at desc").Find(&groups).Error
	return groups, err
}

func FindGroupByID(id uint) (*models.Group, error) {
	var group models.Group
	err := config.DB.
		Preload("Owner").
		Preload("Members").
		Preload("Members.User").
		First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func UpdateGroup(group *models.Group) error {
	return config.DB.Save(group).Error
}

func DeleteGroup(id uint) error {
	// Cascade delete all related data
	config.DB.Where("article_id IN (SELECT id FROM group_articles WHERE group_id = ?)", id).Delete(&models.GroupReaction{})
	config.DB.Where("comment_id IN (SELECT id FROM group_comments WHERE article_id IN (SELECT id FROM group_articles WHERE group_id = ?))", id).Delete(&models.GroupReaction{})
	config.DB.Where("article_id IN (SELECT id FROM group_articles WHERE group_id = ?)", id).Delete(&models.GroupComment{})
	config.DB.Where("group_id = ?", id).Delete(&models.GroupArticle{})
	config.DB.Where("group_id = ?", id).Delete(&models.GroupMember{})
	return config.DB.Delete(&models.Group{}, id).Error
}
