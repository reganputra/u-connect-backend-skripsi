package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type AdminController struct {
	svc service.AdminService
}

func NewAdminController(svc service.AdminService) *AdminController {
	return &AdminController{svc: svc}
}

// ─── Admin helpers ────────────────────────────────────────────────────────────

func getAdminUserID(c *fiber.Ctx) (uint, error) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}
	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fiber.ErrUnauthorized
	}
	id, _ := claims["user_id"].(float64)
	return uint(id), nil
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (ctrl *AdminController) GetDashboard(c *fiber.Ctx) error {
	stats, err := ctrl.svc.GetDashboardStats()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, stats)
}

// ─── User Management ──────────────────────────────────────────────────────────

func (ctrl *AdminController) GetUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	role := c.Query("role")

	users, total, err := ctrl.svc.GetUsers(page, limit, role)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total, "page": page, "limit": limit, "data": users,
	})
}

func (ctrl *AdminController) GetUserByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	u, err := ctrl.svc.GetUserByID(uint(id))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, u)
}

func (ctrl *AdminController) SetUserStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	var req service.UpdateUserStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}
	u, err := ctrl.svc.SetUserStatus(uint(id), req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, u)
}

func (ctrl *AdminController) SetUserRole(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	var req service.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}
	u, err := ctrl.svc.SetUserRole(uint(id), req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, u)
}

// ─── Report Moderation ────────────────────────────────────────────────────────

func (ctrl *AdminController) GetReports(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	status := c.Query("status") // pending | resolved | rejected | ""

	reports, total, err := ctrl.svc.GetReports(page, limit, status)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total, "page": page, "limit": limit, "data": reports,
	})
}

func (ctrl *AdminController) GetReportByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	r, err := ctrl.svc.GetReportByID(uint(id))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, r)
}

func (ctrl *AdminController) ResolveReport(c *fiber.Ctx) error {
	adminID, err := getAdminUserID(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "tidak terautentikasi")
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	var req service.ResolveReportRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}
	r, err := ctrl.svc.ResolveReport(adminID, uint(id), req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, r)
}

func (ctrl *AdminController) RejectReport(c *fiber.Ctx) error {
	adminID, err := getAdminUserID(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "tidak terautentikasi")
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	var req service.RejectReportRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}
	r, err := ctrl.svc.RejectReport(adminID, uint(id), req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, r)
}

// ─── Direct Content Deletion ──────────────────────────────────────────────────

func (ctrl *AdminController) DeletePost(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeletePost(uint(id)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "postingan berhasil dihapus"})
}

func (ctrl *AdminController) DeleteGroup(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeleteGroup(uint(id)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "grup berhasil dihapus"})
}

func (ctrl *AdminController) DeleteEvent(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeleteEvent(uint(id)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "acara berhasil dihapus"})
}

func (ctrl *AdminController) DeleteJob(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeleteJob(uint(id)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "lowongan berhasil dihapus"})
}

// ─── Categories ───────────────────────────────────────────────────────────────

func (ctrl *AdminController) GetCategories(c *fiber.Ctx) error {
	cats, err := ctrl.svc.GetCategories()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, cats)
}

func (ctrl *AdminController) CreateCategory(c *fiber.Ctx) error {
	var req service.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}
	cat, err := ctrl.svc.CreateCategory(req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, cat)
}

func (ctrl *AdminController) UpdateCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	var req service.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}
	cat, err := ctrl.svc.UpdateCategory(uint(id), req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, cat)
}

func (ctrl *AdminController) DeleteCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeleteCategory(uint(id)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "kategori berhasil dihapus"})
}
