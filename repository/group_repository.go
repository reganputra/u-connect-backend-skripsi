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
	// CountGroupStats returns a map of groupID -> [memberCount, articleCount]
	// fetched in a single batch query to avoid N+1 when listing groups.
	CountGroupStats(groupIDs []uint) (map[uint][2]int, error)
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
		Preload("Owner.Profile").
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
		Preload("Owner.Profile").
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
		Preload("Owner.Profile").
		Preload("Members").
		Preload("Members.User").
		Preload("Members.User.Profile").
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

// CountGroupStats returns member and article counts for multiple groups in one
// query, avoiding N+1 when building the groups list response.
func (r *groupRepository) CountGroupStats(groupIDs []uint) (map[uint][2]int, error) {
	if len(groupIDs) == 0 {
		return map[uint][2]int{}, nil
	}

	type statRow struct {
		GroupID      uint `gorm:"column:group_id"`
		MemberCount  int  `gorm:"column:member_count"`
		ArticleCount int  `gorm:"column:article_count"`
	}

	var rows []statRow
	err := r.db.Raw(`
		SELECT
			g.id AS group_id,
			COUNT(DISTINCT gm.id) AS member_count,
			COUNT(DISTINCT ga.id) AS article_count
		FROM groups g
		LEFT JOIN group_members gm
			ON gm.group_id = g.id AND gm.deleted_at IS NULL
		LEFT JOIN group_articles ga
			ON ga.group_id = g.id AND ga.deleted_at IS NULL
		WHERE g.id IN (?)
		GROUP BY g.id
	`, groupIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint][2]int, len(rows))
	for _, row := range rows {
		result[row.GroupID] = [2]int{row.MemberCount, row.ArticleCount}
	}
	return result, nil
}
