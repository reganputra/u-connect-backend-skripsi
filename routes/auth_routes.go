package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterAuthRoutes(app *fiber.App, ctrl *controllers.AuthController) {
	auth := app.Group("/api/auth")
	auth.Post("/register", ctrl.Register)
	auth.Post("/login", ctrl.Login)

	api := app.Group("/api", middleware.Protected())
	api.Get("/me", ctrl.Me)
}
