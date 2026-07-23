package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/dto"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type ReportController struct {
	reportSvc service.ReportService
}

func NewReportController(reportSvc service.ReportService) *ReportController {
	return &ReportController{reportSvc: reportSvc}
}

// CreateReport handles POST /api/reports
// Accessible by: student, alumni, admin
func (ctrl *ReportController) CreateReport(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	var req service.ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}

	report, err := ctrl.reportSvc.CreateReport(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, report)
}

// GetMyReports handles GET /api/reports/mine
// Returns the authenticated user's own submitted reports with pagination
func (ctrl *ReportController) GetMyReports(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	page, limit := utils.ParsePagination(c, 10)

	reports, total, err := ctrl.reportSvc.GetMyReports(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total":   total,
		"page":    page,
		"limit":   limit,
		"reports": reports,
	})
}
