package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
	"github.com/reganputra/skripsi-backend/utils"
)

func RegisterAuthRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")
	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)

	// Protected route example – any authenticated user
	api := app.Group("/api", middleware.Protected())
	api.Get("/me", func(c *fiber.Ctx) error {
		token, ok := c.Locals("user").(*jwt.Token)
		if !ok {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "invalid token")
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "invalid token claims")
		}
		return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
			"user_id": claims["user_id"],
			"email":   claims["email"],
			"role":    claims["role"],
		})
	})
}
