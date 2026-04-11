package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type PortfolioController struct {
	portfolioSvc service.PortfolioService
}

func NewPortfolioController(portfolioSvc service.PortfolioService) *PortfolioController {
	return &PortfolioController{portfolioSvc: portfolioSvc}
}

func (ctrl *PortfolioController) CreatePortfolioItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

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
		Link:        parseOptionalString(c.FormValue("link")),
	}

	item, err := ctrl.portfolioSvc.CreatePortfolioItem(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, item)
}

func (ctrl *PortfolioController) GetPortfolioItems(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	items, err := ctrl.portfolioSvc.GetPortfolioItems(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data portofolio")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, items)
}

func (ctrl *PortfolioController) UpdatePortfolioItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	itemID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID item portofolio tidak valid")
	}

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
		Link:        parseOptionalString(c.FormValue("link")),
	}

	item, err := ctrl.portfolioSvc.UpdatePortfolioItem(userID, uint(itemID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, item)
}

func (ctrl *PortfolioController) DeletePortfolioItem(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	itemID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID item portofolio tidak valid")
	}

	if err := ctrl.portfolioSvc.DeletePortfolioItem(userID, uint(itemID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "item portofolio berhasil dihapus"})
}
