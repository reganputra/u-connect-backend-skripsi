package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterProfileRoutes(app *fiber.App, ctrl *controllers.ProfileController) {
	profile := app.Group("/api/profile", middleware.Protected())

	profile.Post("/", middleware.RequireRole("alumni", "student", "partner"), ctrl.CreateProfile)
	profile.Get("/", ctrl.GetProfile)
	profile.Put("/", middleware.RequireRole("alumni", "student", "partner"), ctrl.UpdateProfile)
	profile.Delete("/", middleware.RequireRole("alumni", "student", "partner"), ctrl.DeleteProfile)

	// Experience sub-routes
	profile.Post("/experience", middleware.RequireRole("alumni", "student", "partner"), ctrl.AddExperience)
	profile.Put("/experience/:id", middleware.RequireRole("alumni", "student", "partner"), ctrl.UpdateExperience)
	profile.Delete("/experience/:id", middleware.RequireRole("alumni", "student", "partner"), ctrl.DeleteExperience)
}
