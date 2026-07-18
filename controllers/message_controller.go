package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type MessageController struct {
	msgSvc service.MessageService
}

func NewMessageController(msgSvc service.MessageService) *MessageController {
	return &MessageController{msgSvc: msgSvc}
}

// GET /api/messages
// Returns one summary entry per conversation partner (last message + unread count).
func (ctrl *MessageController) GetConversationList(c *fiber.Ctx) error {
	callerID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	list, err := ctrl.msgSvc.GetConversationList(callerID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, list)
}

// GET /api/messages/:userID  (paginated, newest-first)
// ?page=1&limit=20
func (ctrl *MessageController) GetConversation(c *fiber.Ctx) error {
	callerID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	partnerID, ok := utils.MustParseIDParam(c, "userID", "pengguna")
	if !ok {
		return nil
	}

	page, limit := utils.ParsePagination(c, 20)

	msgs, total, err := ctrl.msgSvc.GetConversation(callerID, partnerID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total":    total,
		"page":     page,
		"limit":    limit,
		"messages": msgs,
	})
}

// PATCH /api/messages/:userID/read
// Marks all messages from :userID to the caller as read.
func (ctrl *MessageController) MarkAsRead(c *fiber.Ctx) error {
	callerID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	senderID, ok := utils.MustParseIDParam(c, "userID", "pengguna")
	if !ok {
		return nil
	}

	if err := ctrl.msgSvc.MarkAsRead(callerID, senderID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "pesan berhasil ditandai sudah dibaca"})
}

// GET /api/messages/unread
// Returns total unread message count for the caller.
func (ctrl *MessageController) CountUnread(c *fiber.Ctx) error {
	callerID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	count, err := ctrl.msgSvc.CountUnread(callerID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"unread_count": count})
}
