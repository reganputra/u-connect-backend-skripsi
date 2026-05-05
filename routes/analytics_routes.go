package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

// RegisterAnalyticsRoutes registers admin-only analytics endpoints.
// All routes require JWT + admin role.
func RegisterAnalyticsRoutes(app *fiber.App, ctrl *controllers.AnalyticsController) {
	analytics := app.Group("/api/admin/analytics",
		middleware.Protected(),
		middleware.RequireRole("admin"),
	)

	// GET /api/admin/analytics/overview
	analytics.Get("/overview", ctrl.GetOverview)

	// GET /api/admin/analytics/trends?days=30
	analytics.Get("/trends", ctrl.GetTrends)

	// GET /api/admin/analytics/top-content?type=post&days=7&limit=10
	analytics.Get("/top-content", ctrl.GetTopContent)
}
