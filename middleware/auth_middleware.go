package middleware

import (
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/utils"
)

// Protected returns a Fiber middleware that validates Bearer JWT tokens.
func Protected() fiber.Handler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default_secret"
	}
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "missing or invalid token")
		},
	})
}

// RequireRole returns a middleware that allows only the specified roles.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*jwt.Token)
		if !ok {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "missing or invalid token")
		}

		claims, ok := user.Claims.(jwt.MapClaims)
		if !ok {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "invalid token claims")
		}

		role, _ := claims["role"].(string)
		for _, allowed := range roles {
			if role == allowed {
				return c.Next()
			}
		}

		return utils.ErrorResponse(c, fiber.StatusForbidden, "access denied: insufficient role")
	}
}
