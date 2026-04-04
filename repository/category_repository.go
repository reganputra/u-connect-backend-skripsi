package repository

import (
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(c *models.Category) error
	FindAll() ([]models.Category, error)
	FindByID(id uint) (*models.Category, error)
	Update(c *models.Category) error
	Delete(id uint) error
}

type categoryRepository struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(c *models.Category) error {
	return r.db.Create(c).Error
}

func (r *categoryRepository) FindAll() ([]models.Category, error) {
	var cats []models.Category
	err := r.db.Order("name asc").Find(&cats).Error
	return cats, err
}

func (r *categoryRepository) FindByID(id uint) (*models.Category, error) {
	var c models.Category
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Update(c *models.Category) error {
	return r.db.Save(c).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}
