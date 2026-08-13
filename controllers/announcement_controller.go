package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type AnnouncementController struct {
	svc service.AnnouncementService
}

func NewAnnouncementController(svc service.AnnouncementService) *AnnouncementController {
	return &AnnouncementController{svc: svc}
}

func getAdminIDFromCtx(c *fiber.Ctx) (uint, error) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}
	id, _ := claims["user_id"].(float64)
	return uint(id), nil
}

func getUserRoleFromCtx(c *fiber.Ctx) string {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return ""
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	role, _ := claims["role"].(string)
	return role
}

func (ctrl *AnnouncementController) CreateBroadcast(c *fiber.Ctx) error {
	adminID, err := getAdminIDFromCtx(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "tidak terautentikasi")
	}

	var req service.CreateAnnouncementRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}

	a, err := ctrl.svc.CreateBroadcast(adminID, req, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, a)
}

func (ctrl *AnnouncementController) GetAnnouncements(c *fiber.Ctx) error {
	page, limit := utils.ParsePagination(c, 20)
	items, total, err := ctrl.svc.GetAnnouncements(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.PaginatedResponse(c, fiber.StatusOK, items, total, page, limit)
}

func (ctrl *AnnouncementController) DeleteAnnouncement(c *fiber.Ctx) error {
	adminID, err := getAdminIDFromCtx(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "tidak terautentikasi")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id pengumuman tidak valid")
	}

	if err := ctrl.svc.DeleteAnnouncement(adminID, uint(id), c.IP(), c.Get("User-Agent")); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "pengumuman berhasil dihapus"})
}

func (ctrl *AnnouncementController) GetActiveBanners(c *fiber.Ctx) error {
	role := getUserRoleFromCtx(c)
	items, err := ctrl.svc.GetActiveBanners(role)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, items)
}
