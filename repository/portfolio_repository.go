package repository

import (
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/models"
)

func CreatePortfolioItem(item *models.PortfolioItem) error {
	return config.DB.Create(item).Error
}

func FindPortfolioItemsByUserID(userID uint) ([]models.PortfolioItem, error) {
	var items []models.PortfolioItem
	if err := config.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func FindPortfolioItemByID(id uint) (*models.PortfolioItem, error) {
	var item models.PortfolioItem
	if err := config.DB.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func UpdatePortfolioItem(item *models.PortfolioItem) error {
	return config.DB.Save(item).Error
}

func DeletePortfolioItem(id uint) error {
	return config.DB.Delete(&models.PortfolioItem{}, id).Error
}
