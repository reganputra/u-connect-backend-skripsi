package utils

import "github.com/gofiber/fiber/v2"

func SuccessResponse(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func ErrorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

// PaginatedResponse writes the standard success envelope with a paginated
// "data" key plus total/page/limit metadata. It replaces the hand-built
// fiber.Map{"total","page","limit","data":...} envelopes repeated across
// controllers.
func PaginatedResponse(c *fiber.Ctx, status int, data any, total int64, page, limit int) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"data":    data,
	})
}
