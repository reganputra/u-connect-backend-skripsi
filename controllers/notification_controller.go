package controllers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
	"gorm.io/gorm"
)

type NotificationController struct {
	notifSvc service.NotificationService
}

func NewNotificationController(notifSvc service.NotificationService) *NotificationController {
	return &NotificationController{notifSvc: notifSvc}
}

// GET /api/notifications?page=1&limit=20
func (ctrl *NotificationController) GetMyNotifications(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	page, limit := utils.ParsePagination(c, 20)

	notifs, total, err := ctrl.notifSvc.GetMyNotifications(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total":         total,
		"page":          page,
		"limit":         limit,
		"notifications": notifs,
	})
}

// GET /api/notifications/unread
func (ctrl *NotificationController) CountUnread(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	count, err := ctrl.notifSvc.CountUnread(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"unread_count": count})
}

// PATCH /api/notifications/:id/read
func (ctrl *NotificationController) MarkAsRead(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	notifID, ok := utils.MustParseIDParam(c, "id", "notifikasi")
	if !ok {
		return nil
	}
	if err := ctrl.notifSvc.MarkAsRead(uint(notifID), userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "notifikasi tidak ditemukan")
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "notifikasi berhasil ditandai sudah dibaca"})
}

// PATCH /api/notifications/read-all
func (ctrl *NotificationController) MarkAllAsRead(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	if err := ctrl.notifSvc.MarkAllAsRead(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "semua notifikasi ditandai sudah dibaca"})
}
