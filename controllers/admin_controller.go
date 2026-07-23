package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type AdminController struct {
	svc        service.AdminService
	profileSvc service.ProfileService
}

func NewAdminController(svc service.AdminService, profileSvc service.ProfileService) *AdminController {
	return &AdminController{svc: svc, profileSvc: profileSvc}
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
	page, limit := utils.ParsePagination(c, 20)
	role := c.Query("role")
	search := c.Query("search") // cari berdasarkan nama atau email

	users, total, err := ctrl.svc.GetUsers(page, limit, role, search)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.PaginatedResponse(c, fiber.StatusOK, users, total, page, limit)
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
	adminID, _ := getAdminUserID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	var req service.UpdateUserStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}
	u, err := ctrl.svc.SetUserStatus(adminID, uint(id), req, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, u)
}

func (ctrl *AdminController) SetUserRole(c *fiber.Ctx) error {
	adminID, _ := getAdminUserID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	var req service.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "body permintaan tidak valid")
	}
	u, err := ctrl.svc.SetUserRole(adminID, uint(id), req, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, u)
}

// ─── Report Moderation ────────────────────────────────────────────────────────

func (ctrl *AdminController) GetReports(c *fiber.Ctx) error {
	page, limit := utils.ParsePagination(c, 10)
	status := c.Query("status") // pending | resolved | rejected | ""

	reports, total, err := ctrl.svc.GetReports(page, limit, status)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.PaginatedResponse(c, fiber.StatusOK, reports, total, page, limit)
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
	r, err := ctrl.svc.ResolveReport(adminID, uint(id), req, c.IP(), c.Get("User-Agent"))
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
	r, err := ctrl.svc.RejectReport(adminID, uint(id), req, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, r)
}

// ─── Direct Content Deletion ──────────────────────────────────────────────────

func (ctrl *AdminController) DeletePost(c *fiber.Ctx) error {
	adminID, _ := getAdminUserID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeletePost(adminID, uint(id), c.IP(), c.Get("User-Agent")); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "postingan berhasil dihapus"})
}

func (ctrl *AdminController) DeleteGroup(c *fiber.Ctx) error {
	adminID, _ := getAdminUserID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeleteGroup(adminID, uint(id), c.IP(), c.Get("User-Agent")); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "grup berhasil dihapus"})
}

func (ctrl *AdminController) DeleteEvent(c *fiber.Ctx) error {
	adminID, _ := getAdminUserID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeleteEvent(adminID, uint(id), c.IP(), c.Get("User-Agent")); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "acara berhasil dihapus"})
}

func (ctrl *AdminController) DeleteJob(c *fiber.Ctx) error {
	adminID, _ := getAdminUserID(c)
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}
	if err := ctrl.svc.DeleteJob(adminID, uint(id), c.IP(), c.Get("User-Agent")); err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "lowongan berhasil dihapus"})
}

// ─── Activity Logs ────────────────────────────────────────────────────────────

func (ctrl *AdminController) GetAdminLogs(c *fiber.Ctx) error {
	page, limit := utils.ParsePagination(c, 20)
	action := c.Query("action")
	targetType := c.Query("target_type")

	logs, total, err := ctrl.svc.GetAdminActivityLogs(page, limit, action, targetType)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.PaginatedResponse(c, fiber.StatusOK, logs, total, page, limit)
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

// PATCH /api/admin/users/:id/profile
func (ctrl *AdminController) PatchUserProfile(c *fiber.Ctx) error {

	targetID, ok := utils.MustParseIDParam(c, "id", "pengguna")
	if !ok {
		return nil
	}
	var req service.AdminProfilePatchRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "request body tidak valid")
	}
	if err := ctrl.profileSvc.AdminPatchProfile(targetID, req); err != nil {
		if err.Error() == "tidak ada field yang diberikan untuk diupdate" {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "profil pengguna berhasil diperbarui"})
}

// POST /api/admin/users/:id/experience
func (ctrl *AdminController) AddUserExperience(c *fiber.Ctx) error {

	targetID, ok := utils.MustParseIDParam(c, "id", "pengguna")
	if !ok {
		return nil
	}
	var req service.ExperienceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "request body tidak valid")
	}
	exp, err := ctrl.profileSvc.AddExperience(targetID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, exp)
}

// PUT /api/admin/users/:id/experience/:expId
func (ctrl *AdminController) UpdateUserExperience(c *fiber.Ctx) error {

	targetID, ok := utils.MustParseIDParam(c, "id", "pengguna")
	if !ok {
		return nil
	}

	expID, ok := utils.MustParseIDParam(c, "expId", "pengalaman")
	if !ok {
		return nil
	}
	var req service.ExperienceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "request body tidak valid")
	}
	exp, err := ctrl.profileSvc.UpdateExperience(targetID, expID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, exp)
}

// DELETE /api/admin/users/:id/experience/:expId
func (ctrl *AdminController) DeleteUserExperience(c *fiber.Ctx) error {
	targetID, ok := utils.MustParseIDParam(c, "id", "pengguna")
	if !ok {
		return nil
	}
	expID, ok := utils.MustParseIDParam(c, "expId", "pengalaman")
	if !ok {
		return nil
	}
	if err := ctrl.profileSvc.DeleteExperience(targetID, expID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "pengalaman berhasil dihapus"})
}
