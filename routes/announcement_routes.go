package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterAnnouncementRoutes(app *fiber.App, ctrl *controllers.AnnouncementController) {
	// Public (Authenticated): active banners for sticky header display
	app.Get("/api/announcements/active", middleware.Protected(), ctrl.GetActiveBanners)

	// Admin routes
	admin := app.Group("/api/admin/announcements", middleware.Protected(), middleware.RequireRole("admin"))
	admin.Post("/", ctrl.CreateBroadcast)
	admin.Get("/", ctrl.GetAnnouncements)
	admin.Delete("/:id", ctrl.DeleteAnnouncement)
}
