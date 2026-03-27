package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type PortfolioRepository interface {
	CreatePortfolioItem(item *models.PortfolioItem) error
	FindPortfolioItemsByUserID(userID uint) ([]models.PortfolioItem, error)
	FindPortfolioItemByID(id uint) (*models.PortfolioItem, error)
	UpdatePortfolioItem(item *models.PortfolioItem) error
	DeletePortfolioItem(id uint) error
}

type portfolioRepository struct {
	db *gorm.DB
}

func NewPortfolioRepository(db *gorm.DB) PortfolioRepository {
	return &portfolioRepository{db: db}
}

func (r *portfolioRepository) CreatePortfolioItem(item *models.PortfolioItem) error {
	return r.db.Create(item).Error
}

func (r *portfolioRepository) FindPortfolioItemsByUserID(userID uint) ([]models.PortfolioItem, error) {
	var items []models.PortfolioItem
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *portfolioRepository) FindPortfolioItemByID(id uint) (*models.PortfolioItem, error) {
	var item models.PortfolioItem
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *portfolioRepository) UpdatePortfolioItem(item *models.PortfolioItem) error {
	return r.db.Save(item).Error
}

func (r *portfolioRepository) DeletePortfolioItem(id uint) error {
	return r.db.Delete(&models.PortfolioItem{}, id).Error
}
