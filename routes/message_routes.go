package routes

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/ws"
)

func SetupMessageRoutes(
	app *fiber.App,
	ctrl *controllers.MessageController,
	hub *ws.Hub,
	msgSvc service.MessageService,
	userRepo repository.UserRepository,
	notifSvc service.NotificationService,
	tickets *ws.TicketStore,
) {
	// ── REST endpoints ─────────────────────────────────────────────────────────
	msgs := app.Group("/api/messages", middleware.Protected(), middleware.RequireRole("student", "alumni"))

	msgs.Get("/", ctrl.GetConversationList)
	msgs.Get("/unread", ctrl.CountUnread)
	msgs.Get("/:userID", ctrl.GetConversation)
	msgs.Patch("/:userID/read", ctrl.MarkAsRead)

	// ── WebSocket endpoint ─────────────────────────────────────────────────────
	app.Post("/api/ws/ticket", middleware.Protected(), middleware.RequireRole("student", "alumni"), controllers.NewWSTicketController(tickets).CreateTicket)

	// Step 2 — Confirm the request is a WebSocket upgrade, then upgrade.
	app.Get("/api/ws", ws.WSAuthMiddleware(userRepo, tickets), func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			log.Printf("[WS/UPGRADE] WARN rejected — not a WebSocket request — method: %s, ip: %s", c.Method(), c.IP())
			return c.Status(fiber.StatusUpgradeRequired).JSON(fiber.Map{
				"error": "websocket upgrade required",
			})
		}
		log.Printf("[WS/UPGRADE] INFO upgrade request accepted — ip: %s", c.IP())
		return c.Next()
	}, ws.WSHandler(hub, msgSvc, userRepo, notifSvc))
}
