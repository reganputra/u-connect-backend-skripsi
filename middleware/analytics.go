package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// ViewCooldownDuration is the minimum time that must elapse before the same
// authenticated user can generate a second view event for the same piece of
// content. Requests that arrive within this window are silently dropped so
// that repeated page refreshes or frontend re-fetches do not inflate counts.
//
// 1 hour mirrors common industry practice for "unique hourly viewer" metrics
// and is straightforward to explain in academic documentation.
const ViewCooldownDuration = 1 * time.Hour

// TrackView returns a Fiber middleware that appends a PageView row
// asynchronously after the handler has responded.
//
// Rules:
//   - Only records views for GET requests that return HTTP 200.
//   - Extracts the resource ID from the ":id" route parameter.
//   - Captures the authenticated user's ID from the JWT stored in c.Locals("user").
//     If no JWT is present (future public routes), the cooldown check is skipped
//     and the view is recorded with a nil user_id.
//   - The deduplication gate runs inside the goroutine so it never adds latency
//     to the handler's response time.
func TrackView(targetType string, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Run the actual handler first.
		err := c.Next()

		// Only track successful GET requests.
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
			// Deduplication gate — only enforced for authenticated users because
			// anonymous views cannot be attributed to a stable identity.
			if userID != nil {
				cutoff := time.Now().UTC().Add(-ViewCooldownDuration)

				var count int64
				db.Model(&models.PageView{}).
					Where(
						"user_id = ? AND target_type = ? AND target_id = ? AND created_at >= ?",
						*userID, targetType, uint(id), cutoff,
					).
					Limit(1).
					Count(&count)

				if count > 0 {
					// View already recorded within the cooldown window — skip insert.
					return
				}
			}

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
// Returns nil if no token is present or the token is malformed (e.g. public routes).
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
