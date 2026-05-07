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
	auth.Post("/forgot-password", ctrl.ForgotPassword) // No auth required

	api := app.Group("/api")
	api.Get("/me", middleware.Protected(), ctrl.Me)
	api.Post("/auth/change-password", middleware.Protected(), ctrl.ChangePassword)
}
