package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterReportRoutes(app *fiber.App, ctrl *controllers.ReportController) {
	reports := app.Group("/api/reports", middleware.Protected())

	// Submit a report — student, alumni, and admin can report content
	reports.Post("/", middleware.RequireRole("student", "alumni", "admin"), ctrl.CreateReport)

	// View own submitted reports
	reports.Get("/mine", middleware.RequireRole("student", "alumni", "admin"), ctrl.GetMyReports)
}
