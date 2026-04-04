package routes

import (
	fws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/ws"
)

func SetupMessageRoutes(app *fiber.App, ctrl *controllers.MessageController, hub *ws.Hub, msgSvc service.MessageService) {
	// ── REST endpoints ─────────────────────────────────────────────────────────
	msgs := app.Group("/api/messages", middleware.Protected(), middleware.RequireRole("student", "alumni"))

	msgs.Get("/", ctrl.GetConversationList)
	msgs.Get("/unread", ctrl.CountUnread)
	msgs.Get("/:userID", ctrl.GetConversation)
	msgs.Patch("/:userID/read", ctrl.MarkAsRead)

	// WebSocket endpoint — auth done inside handler via ?token=<jwt>
	app.Get("/api/ws", func(c *fiber.Ctx) error {
		if fws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}, ws.WSHandler(hub, msgSvc))

}
