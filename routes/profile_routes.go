package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterProfileRoutes(app *fiber.App) {
	profile := app.Group("/api/profile", middleware.Protected())

	profile.Post("/", controllers.CreateProfile)
	profile.Get("/", controllers.GetProfile)
	profile.Put("/", controllers.UpdateProfile)
	profile.Delete("/", controllers.DeleteProfile)
	profile.Post("/picture", controllers.UploadProfilePicture)

	// Experience sub-routes
	profile.Post("/experience", controllers.AddExperience)
	profile.Put("/experience/:id", controllers.UpdateExperience)
	profile.Delete("/experience/:id", controllers.DeleteExperience)
}
