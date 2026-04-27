package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type GroupRepository interface {
	CreateGroup(group *models.Group) error
	FindGroups(page, limit int) ([]models.Group, int64, error)
	FindGroupsByOwner(ownerID uint, page, limit int) ([]models.Group, int64, error)
	FindGroupByID(id uint) (*models.Group, error)
	UpdateGroup(group *models.Group) error
	DeleteGroup(id uint) error
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) CreateGroup(group *models.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepository) FindGroups(page, limit int) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64

	if err := r.db.Model(&models.Group{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := r.db.Preload("Owner").
		Order("created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&groups).Error
	return groups, total, err
}

func (r *groupRepository) FindGroupsByOwner(ownerID uint, page, limit int) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64

	base := r.db.Model(&models.Group{}).Where("owner_id = ?", ownerID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := r.db.
		Preload("Owner").
		Where("owner_id = ?", ownerID).
		Order("created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&groups).Error

	return groups, total, err
}

func (r *groupRepository) FindGroupByID(id uint) (*models.Group, error) {
	var group models.Group
	err := r.db.
		Preload("Owner").
		Preload("Members").
		Preload("Members.User").
		Preload("Articles").
		Preload("Articles.User").
		First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) UpdateGroup(group *models.Group) error {
	return r.db.Save(group).Error
}

func (r *groupRepository) DeleteGroup(id uint) error {
	r.db.Where("article_id IN (SELECT id FROM group_articles WHERE group_id = ?)", id).Delete(&models.GroupReaction{})
	r.db.Where("comment_id IN (SELECT id FROM group_comments WHERE article_id IN (SELECT id FROM group_articles WHERE group_id = ?))", id).Delete(&models.GroupReaction{})
	r.db.Where("article_id IN (SELECT id FROM group_articles WHERE group_id = ?)", id).Delete(&models.GroupComment{})
	r.db.Where("group_id = ?", id).Delete(&models.GroupArticle{})
	r.db.Where("group_id = ?", id).Delete(&models.GroupMember{})
	return r.db.Delete(&models.Group{}, id).Error
}
