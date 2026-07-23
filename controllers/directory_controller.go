package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type DirectoryController struct {
	profileSvc   service.ProfileService
	portfolioSvc service.PortfolioService
}

func NewDirectoryController(profileSvc service.ProfileService, portfolioSvc service.PortfolioService) *DirectoryController {
	return &DirectoryController{profileSvc: profileSvc, portfolioSvc: portfolioSvc}
}

// Mengembalikan profil publik lengkap untuk pengguna tertentu.
func (ctrl *DirectoryController) GetPublicProfile(c *fiber.Ctx) error {
	userID, ok := utils.MustParseIDParam(c, "userID", "pengguna")
	if !ok {
		return nil
	}

	profile, err := ctrl.profileSvc.GetPublicProfile(uint(userID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

// Mengembalikan item portofolio publik untuk pengguna tertentu.
func (ctrl *DirectoryController) GetPublicPortfolio(c *fiber.Ctx) error {
	userID, ok := utils.MustParseIDParam(c, "userID", "pengguna")
	if !ok {
		return nil
	}

	portfolioItems, err := ctrl.portfolioSvc.GetPublicPortfolioItems(uint(userID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data portofolio")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, portfolioItems)
}

// Mengembalikan daftar profil pengguna yang dipaginasi (mahasiswa + alumni).
func (ctrl *DirectoryController) GetDirectory(c *fiber.Ctx) error {
	page, limit := utils.ParsePagination(c, 20)

	profiles, total, err := ctrl.profileSvc.GetProfileDirectory(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.PaginatedResponse(c, fiber.StatusOK, profiles, total, page, limit)
}

// Mencari profil berdasarkan nama, keahlian, perusahaan, atau minat.
func (ctrl *DirectoryController) SearchProfiles(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "parameter q (query) diperlukan")
	}

	page, limit := utils.ParsePagination(c, 20)

	profiles, total, err := ctrl.profileSvc.SearchProfiles(query, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"query": query,
		"data":  profiles,
	})
}

// Mengembalikan profil yang dipaginasi yang difilter berdasarkan peran (student, alumni, atau partner).
func (ctrl *DirectoryController) GetProfilesByRole(c *fiber.Ctx) error {
	role := c.Params("role")
	if role != "student" && role != "alumni" && role != "partner" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "peran harus student, alumni, atau partner")
	}

	page, limit := utils.ParsePagination(c, 20)

	profiles, total, err := ctrl.profileSvc.GetProfilesByRole(role, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"role":  role,
		"data":  profiles,
	})
}
