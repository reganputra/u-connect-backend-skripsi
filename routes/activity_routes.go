package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterActivityRoutes(app *fiber.App, ctrl *controllers.ActivityController) {
	activity := app.Group("/api/me/activity", middleware.Protected())
	activity.Get("/summary", ctrl.GetMyActivitySummary)
}
