package controllers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/utils"
	"github.com/reganputra/skripsi-backend/ws"
	"strings"
	"time"
)

type WSTicketController struct{ tickets *ws.TicketStore }

func NewWSTicketController(tickets *ws.TicketStore) *WSTicketController {
	return &WSTicketController{tickets: tickets}
}

// CreateTicket is mounted behind Protected and RequireRole. It never echoes a JWT.
func (ctrl *WSTicketController) CreateTicket(c *fiber.Ctx) error {
	raw := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	if _, err := utils.ParseJWT(raw); err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()
	ticket, err := ctrl.tickets.Issue(ctx, raw)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusServiceUnavailable, "ticket storage unavailable")
	}
	c.Set("Cache-Control", "no-store")
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"ticket": ticket, "ttl": 30})
}
