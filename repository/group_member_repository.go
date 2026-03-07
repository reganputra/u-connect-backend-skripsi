package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func AddGroupMember(member *models.GroupMember) error {
	return config.DB.Create(member).Error
}

func FindGroupMember(groupID, userID uint) (*models.GroupMember, error) {
	var member models.GroupMember
	err := config.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func FindGroupMembers(groupID uint) ([]models.GroupMember, error) {
	var members []models.GroupMember
	err := config.DB.Where("group_id = ?", groupID).Preload("User").Find(&members).Error
	return members, err
}

func FindJoinedGroups(userID uint) ([]models.Group, error) {
	var groups []models.Group
	err := config.DB.
		Joins("JOIN group_members ON group_members.group_id = groups.id AND group_members.deleted_at IS NULL").
		Where("group_members.user_id = ?", userID).
		Preload("Owner").
		Order("groups.created_at desc").
		Find(&groups).Error
	return groups, err
}

func RemoveGroupMember(groupID, userID uint) error {
	return config.DB.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

func CountGroupMembers(groupID uint) (int64, error) {
	var count int64
	err := config.DB.Model(&models.GroupMember{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}
