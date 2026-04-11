package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func SetupDirectoryRoutes(app *fiber.App, ctrl *controllers.DirectoryController) {
	// Public directory browsing — available to student, alumni, and partner
	dir := app.Group("/api/directory", middleware.Protected(), middleware.RequireRole("student", "alumni", "partner"))

	dir.Get("/", ctrl.GetDirectory)
	dir.Get("/search", ctrl.SearchProfiles)
	dir.Get("/role/:role", ctrl.GetProfilesByRole)
	dir.Get("/:userID/portfolio", ctrl.GetPublicPortfolio)
	dir.Get("/:userID", ctrl.GetPublicProfile)
}
