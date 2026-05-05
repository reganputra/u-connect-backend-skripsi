package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type AnalyticsController struct {
	svc service.AnalyticsService
}

func NewAnalyticsController(svc service.AnalyticsService) *AnalyticsController {
	return &AnalyticsController{svc: svc}
}

// Svc exposes the underlying AnalyticsService for use by the scheduler.
func (ctrl *AnalyticsController) Svc() service.AnalyticsService {
	return ctrl.svc
}

// GET /api/admin/analytics/overview
// Returns lifetime enhanced stats + yesterday's daily snapshot.
func (ctrl *AnalyticsController) GetOverview(c *fiber.Ctx) error {
	overview, err := ctrl.svc.GetOverview()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data analitik")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, overview)
}

// GET /api/admin/analytics/trends?days=30
// Returns an array of daily snapshots for the last `days` days (max 365).
func (ctrl *AnalyticsController) GetTrends(c *fiber.Ctx) error {
	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	snaps, err := ctrl.svc.GetTrends(days)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data tren")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"days": days,
		"data": snaps,
	})
}

// GET /api/admin/analytics/top-content?type=post&days=7&limit=10
// Returns the top N content items by view count in the last `days` days.
// `type` must be one of: post | group.
func (ctrl *AnalyticsController) GetTopContent(c *fiber.Ctx) error {
	targetType := c.Query("type", "post")
	validTypes := map[string]bool{"post": true, "group": true}
	if !validTypes[targetType] {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "type harus salah satu dari: post, group")
	}

	days, _ := strconv.Atoi(c.Query("days", "7"))
	if days < 1 || days > 365 {
		days = 7
	}
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	items, err := ctrl.svc.GetTopContent(targetType, days, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data konten teratas")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"type":  targetType,
		"days":  days,
		"limit": limit,
		"data":  items,
	})
}
