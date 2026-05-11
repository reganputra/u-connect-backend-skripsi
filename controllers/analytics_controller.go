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

// Svc mengekspos service AnalyticsService yang mendasarinya agar dapat digunakan oleh scheduler.
func (ctrl *AnalyticsController) Svc() service.AnalyticsService {
	return ctrl.svc
}

// Mengembalikan statistik lengkap + snapshot harian kemarin.
func (ctrl *AnalyticsController) GetOverview(c *fiber.Ctx) error {
	overview, err := ctrl.svc.GetOverview()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data analitik")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, overview)
}

// Mengembalikan array snapshot harian untuk `days` hari terakhir (maks 365).
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

// Mengembalikan item konten teratas berdasarkan jumlah tayangan dalam `days` hari terakhir.
// `type` harus salah satu dari: post | group.
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
