package repository

import (
	"errors"
	"time"

	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type GroupMemberRepository interface {
	AddGroupMember(member *models.GroupMember) error
	FindGroupMember(groupID, userID uint) (*models.GroupMember, error)
	FindGroupMembers(groupID uint, page, limit int) ([]models.GroupMember, int64, error)
	FindJoinedGroups(userID uint, page, limit int) ([]models.Group, int64, error)
	RemoveGroupMember(groupID, userID uint) error
	CountGroupMembers(groupID uint) (int64, error)
}

type groupMemberRepository struct {
	db *gorm.DB
}

func NewGroupMemberRepository(db *gorm.DB) GroupMemberRepository {
	return &groupMemberRepository{db: db}
}

func (r *groupMemberRepository) AddGroupMember(member *models.GroupMember) error {
	// Handle re-join gracefully when a previous membership exists as soft-deleted.
	var existing models.GroupMember
	err := r.db.Unscoped().
		Where("group_id = ? AND user_id = ?", member.GroupID, member.UserID).
		First(&existing).Error
	if err == nil {
		if existing.DeletedAt.Valid {
			return r.db.Unscoped().Model(&models.GroupMember{}).
				Where("id = ?", existing.ID).
				Updates(map[string]interface{}{
					"deleted_at": nil,
					"role":       member.Role,
					"updated_at": time.Now(),
				}).Error
		}
		return errors.New("sudah menjadi anggota grup ini")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return r.db.Create(member).Error
}

func (r *groupMemberRepository) FindGroupMember(groupID, userID uint) (*models.GroupMember, error) {
	var member models.GroupMember
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *groupMemberRepository) FindGroupMembers(groupID uint, page, limit int) ([]models.GroupMember, int64, error) {
	var members []models.GroupMember
	var total int64

	q := r.db.Model(&models.GroupMember{}).Where("group_id = ?", groupID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := r.db.Where("group_id = ?", groupID).
		Preload("User").
		Order("created_at asc").
		Offset(offset).
		Limit(limit).
		Find(&members).Error
	return members, total, err
}

func (r *groupMemberRepository) FindJoinedGroups(userID uint, page, limit int) ([]models.Group, int64, error) {
	var groups []models.Group
	q := r.db.Model(&models.Group{}).
		Joins("JOIN group_members ON group_members.group_id = groups.id AND group_members.deleted_at IS NULL").
		Where("group_members.user_id = ?", userID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := q.
		Preload("Owner").
		Order("groups.created_at desc").
		Offset(offset).
		Limit(limit).
		Find(&groups).Error
	return groups, total, err
}

func (r *groupMemberRepository) RemoveGroupMember(groupID, userID uint) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

func (r *groupMemberRepository) CountGroupMembers(groupID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.GroupMember{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}
