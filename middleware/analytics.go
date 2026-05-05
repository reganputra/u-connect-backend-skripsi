package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// TrackView returns a Fiber middleware that appends a PageView row asynchronously
// after the handler responds.
//
// Rules:
//   - Only records views for GET requests that return HTTP 200.
//   - Extracts the resource ID from the ":id" route parameter.
//   - Captures the authenticated user's ID from the JWT in c.Locals("user").
//     If no JWT is present (future public routes), user_id is stored as nil.
//   - The database insert happens in a fire-and-forget goroutine so it never
//     adds latency to the handler's response time.
func TrackView(targetType string, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Run the real handler first.
		err := c.Next()

		// Only record views for successful GET requests.
		if c.Method() != "GET" || c.Response().StatusCode() != 200 {
			return err
		}

		idStr := c.Params("id")
		id, parseErr := strconv.ParseUint(idStr, 10, 64)
		if parseErr != nil || id == 0 {
			return err
		}

		userID := extractViewUserID(c)

		go func() {
			_ = db.Create(&models.PageView{
				CreatedAt:  time.Now().UTC(),
				UserID:     userID,
				TargetType: targetType,
				TargetID:   uint(id),
			}).Error
		}()

		return err
	}
}

// extractViewUserID extracts the user_id from the JWT stored in c.Locals("user").
// Returns nil if the token is absent or malformed (e.g. for unauthenticated routes).
func extractViewUserID(c *fiber.Ctx) *uint {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	idFloat, ok := claims["user_id"].(float64)
	if !ok || idFloat <= 0 {
		return nil
	}
	id := uint(idFloat)
	return &id
}
