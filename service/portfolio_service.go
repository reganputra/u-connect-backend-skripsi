package service

import (
	"errors"

	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
)

type PortfolioItemRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Tags        *string `json:"tags"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	MediaURL    *string `json:"media_url"` // set by controller after Cloudinary upload
}

func CreatePortfolioItem(userID uint, req PortfolioItemRequest) (*models.PortfolioItem, error) {
	if req.Title == "" {
		return nil, errors.New("judul wajib diisi")
	}

	item := &models.PortfolioItem{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		MediaURL:    req.MediaURL,
	}

	if err := repository.CreatePortfolioItem(item); err != nil {
		return nil, errors.New("gagal membuat item portofolio")
	}
	return item, nil
}

func GetPortfolioItems(userID uint) ([]models.PortfolioItem, error) {
	return repository.FindPortfolioItemsByUserID(userID)
}

func UpdatePortfolioItem(userID uint, itemID uint, req PortfolioItemRequest) (*models.PortfolioItem, error) {
	item, err := repository.FindPortfolioItemByID(itemID)
	if err != nil {
		return nil, errors.New("item portofolio tidak ditemukan")
	}

	// Ownership check
	if item.UserID != userID {
		return nil, errors.New("akses ditolak")
	}

	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Description != nil {
		item.Description = req.Description
	}
	if req.Category != nil {
		item.Category = req.Category
	}
	if req.Tags != nil {
		item.Tags = req.Tags
	}
	if req.StartDate != nil {
		item.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		item.EndDate = req.EndDate
	}
	if req.MediaURL != nil {
		item.MediaURL = req.MediaURL
	}

	if err := repository.UpdatePortfolioItem(item); err != nil {
		return nil, errors.New("gagal memperbarui item portofolio")
	}
	return item, nil
}

func DeletePortfolioItem(userID uint, itemID uint) error {
	item, err := repository.FindPortfolioItemByID(itemID)
	if err != nil {
		return errors.New("item portofolio tidak ditemukan")
	}
	if item.UserID != userID {
		return errors.New("akses ditolak")
	}
	return repository.DeletePortfolioItem(itemID)
}

func UpdatePortfolioMedia(userID uint, itemID uint, mediaURL string) error {
	item, err := repository.FindPortfolioItemByID(itemID)
	if err != nil {
		return errors.New("item portofolio tidak ditemukan")
	}
	if item.UserID != userID {
		return errors.New("akses ditolak")
	}
	item.MediaURL = &mediaURL
	return repository.UpdatePortfolioItem(item)
}
