package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func SetupNotificationRoutes(app *fiber.App, ctrl *controllers.NotificationController) {
	notifs := app.Group("/api/notifications", middleware.Protected())

	// NOTE: /read-all must be registered before /:id/read to avoid Fiber
	// interpreting "read-all" as an :id parameter.
	notifs.Get("/", ctrl.GetMyNotifications)
	notifs.Get("/unread", ctrl.CountUnread)
	notifs.Patch("/read-all", ctrl.MarkAllAsRead)
	notifs.Patch("/:id/read", ctrl.MarkAsRead)
}
