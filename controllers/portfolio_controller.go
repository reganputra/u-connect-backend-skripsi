package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

func CreatePortfolioItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	// Handle optional media upload
	mediaURL, err := uploadFileIfPresent(c, "media", "alumni-platform/portfolio")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	req := service.PortfolioItemRequest{
		Title:       c.FormValue("title"),
		Description: parseOptionalString(c.FormValue("description")),
		Category:    parseOptionalString(c.FormValue("category")),
		Tags:        parseOptionalString(c.FormValue("tags")),
		StartDate:   parseOptionalString(c.FormValue("start_date")),
		EndDate:     parseOptionalString(c.FormValue("end_date")),
		MediaURL:    parseOptionalString(mediaURL),
	}

	item, err := service.CreatePortfolioItem(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, item)
}

func GetPortfolioItems(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	items, err := service.GetPortfolioItems(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data portofolio")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, items)
}

func UpdatePortfolioItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	itemID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID item portofolio tidak valid")
	}

	// Handle optional media upload
	mediaURL, err := uploadFileIfPresent(c, "media", "alumni-platform/portfolio")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	req := service.PortfolioItemRequest{
		Title:       c.FormValue("title"),
		Description: parseOptionalString(c.FormValue("description")),
		Category:    parseOptionalString(c.FormValue("category")),
		Tags:        parseOptionalString(c.FormValue("tags")),
		StartDate:   parseOptionalString(c.FormValue("start_date")),
		EndDate:     parseOptionalString(c.FormValue("end_date")),
		MediaURL:    parseOptionalString(mediaURL),
	}

	item, err := service.UpdatePortfolioItem(userID, uint(itemID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, item)
}

func DeletePortfolioItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	itemID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID item portofolio tidak valid")
	}

	if err := service.DeletePortfolioItem(userID, uint(itemID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "item portofolio berhasil dihapus"})
}
