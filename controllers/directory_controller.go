package controllers

import (
	"strconv"

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

// GET /api/directory/:userID
// Returns detailed public profile for a specific user.
func (ctrl *DirectoryController) GetPublicProfile(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("userID"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID pengguna tidak valid")
	}

	profile, err := ctrl.profileSvc.GetPublicProfile(uint(userID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	// Keep this response shape identical to GET /api/profile
	// so frontend can reuse the same profile detail renderer.
	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

// GET /api/directory/:userID/portfolio
// Returns public portfolio items for a specific user.
func (ctrl *DirectoryController) GetPublicPortfolio(c *fiber.Ctx) error {
	userID, err := strconv.ParseUint(c.Params("userID"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID pengguna tidak valid")
	}

	portfolioItems, err := ctrl.portfolioSvc.GetPublicPortfolioItems(uint(userID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data portofolio")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, portfolioItems)
}

// GET /api/directory
// Returns paginated list of all user profiles (students + alumni).
func (ctrl *DirectoryController) GetDirectory(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	profiles, total, err := ctrl.profileSvc.GetProfileDirectory(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  profiles,
	})
}

// GET /api/directory/search?q=<query>
// Searches profiles by name, skills, company, or interests.
func (ctrl *DirectoryController) SearchProfiles(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "parameter q (query) diperlukan")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

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

// GET /api/directory/role/:role
// Returns paginated profiles filtered by role (student or alumni).
func (ctrl *DirectoryController) GetProfilesByRole(c *fiber.Ctx) error {
	role := c.Params("role")
	if role != "student" && role != "alumni" {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "peran harus student atau alumni")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

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
