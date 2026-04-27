package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type GroupController struct {
	groupSvc service.GroupService
}

func NewGroupController(groupSvc service.GroupService) *GroupController {
	return &GroupController{groupSvc: groupSvc}
}

// ─── Group CRUD ───────────────────────────────────────────────────────────────

func (ctrl *GroupController) CreateGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	bannerURL, err := uploadFileIfPresent(c, "banner", "alumni-platform/groups/banners")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	req := service.GroupRequest{
		Category:    c.FormValue("category"),
		Title:       c.FormValue("title"),
		Description: parseOptionalString(c.FormValue("description")),
		Rules:       parseOptionalString(c.FormValue("rules")),
		BannerURL:   parseOptionalString(bannerURL),
	}
	group, err := ctrl.groupSvc.CreateGroup(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, group)
}

func (ctrl *GroupController) GetGroups(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	groups, total, err := ctrl.groupSvc.GetGroups(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data grup")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  groups,
	})
}

func (ctrl *GroupController) GetOwnedGroups(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	groups, total, err := ctrl.groupSvc.GetOwnedGroups(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data grup milik pengguna")
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  groups,
	})
}

func (ctrl *GroupController) GetGroupByID(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	group, err := ctrl.groupSvc.GetGroupByID(uint(groupID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, group)
}

func (ctrl *GroupController) UpdateGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	bannerURL, err := uploadFileIfPresent(c, "banner", "alumni-platform/groups/banners")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	req := service.GroupRequest{
		Category:    c.FormValue("category"),
		Title:       c.FormValue("title"),
		Description: parseOptionalString(c.FormValue("description")),
		Rules:       parseOptionalString(c.FormValue("rules")),
		BannerURL:   parseOptionalString(bannerURL),
	}
	group, err := ctrl.groupSvc.UpdateGroup(userID, uint(groupID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya pemilik grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, group)
}

func (ctrl *GroupController) DeleteGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	if err := ctrl.groupSvc.DeleteGroup(userID, uint(groupID)); err != nil {
		if err.Error() == "akses ditolak: hanya pemilik grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "grup berhasil dihapus"})
}

// ─── Membership ───────────────────────────────────────────────────────────────

func (ctrl *GroupController) JoinGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	if err := ctrl.groupSvc.JoinGroup(userID, uint(groupID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil bergabung dengan grup"})
}

func (ctrl *GroupController) LeaveGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	if err := ctrl.groupSvc.LeaveGroup(userID, uint(groupID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil meninggalkan grup"})
}

func (ctrl *GroupController) GetGroupMembers(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	members, total, err := ctrl.groupSvc.GetGroupMembers(uint(groupID), page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data anggota")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total":   total,
		"page":    page,
		"limit":   limit,
		"members": members,
	})
}

func (ctrl *GroupController) GetJoinedGroups(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	groups, total, err := ctrl.groupSvc.GetJoinedGroups(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data grup yang diikuti")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  groups,
	})
}

func (ctrl *GroupController) KickMember(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	targetID, err := strconv.ParseUint(c.Params("userID"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID pengguna tidak valid")
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.BodyParser(&body)
	if err := ctrl.groupSvc.KickMember(userID, uint(groupID), uint(targetID), body.Reason); err != nil {
		if err.Error() == "akses ditolak: hanya pemilik grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "anggota berhasil dikeluarkan"})
}

// ─── Articles ─────────────────────────────────────────────────────────────────

func (ctrl *GroupController) CreateGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}

	mediaURLs, err := uploadFilesIfPresent(c, "medias", "alumni-platform/groups/articles")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	// Also support legacy single "media" field for backward compatibility
	if len(mediaURLs) == 0 {
		mediaURL, err := uploadFileIfPresent(c, "media", "alumni-platform/groups/articles")
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		if mediaURL != "" {
			mediaURLs = []string{mediaURL}
		}
	}

	req := service.GroupArticleRequest{
		Title:     c.FormValue("title"),
		Content:   c.FormValue("content"),
		MediaURLs: mediaURLs,
	}
	article, err := ctrl.groupSvc.CreateGroupArticle(userID, uint(groupID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, article)
}

func (ctrl *GroupController) GetGroupArticleDetail(c *fiber.Ctx) error {
	userID, _ := getUserIDFromToken(c)
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID artikel tidak valid")
	}
	detail, err := ctrl.groupSvc.GetGroupArticleDetail(userID, uint(articleID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, detail)
}

func (ctrl *GroupController) UpdateGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID artikel tidak valid")
	}

	mediaURLs, err := uploadFilesIfPresent(c, "medias", "alumni-platform/groups/articles")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	// Also support legacy single "media" field for backward compatibility
	if len(mediaURLs) == 0 {
		mediaURL, err := uploadFileIfPresent(c, "media", "alumni-platform/groups/articles")
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		if mediaURL != "" {
			mediaURLs = []string{mediaURL}
		}
	}

	req := service.GroupArticleRequest{
		Title:     c.FormValue("title"),
		Content:   c.FormValue("content"),
		MediaURLs: mediaURLs,
	}
	article, err := ctrl.groupSvc.UpdateGroupArticle(userID, uint(articleID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, article)
}

func (ctrl *GroupController) DeleteGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID artikel tidak valid")
	}
	if err := ctrl.groupSvc.DeleteGroupArticle(userID, uint(articleID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "artikel berhasil dihapus"})
}

// ─── Comments ─────────────────────────────────────────────────────────────────

func (ctrl *GroupController) AddGroupComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID artikel tidak valid")
	}
	var req service.GroupCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	comment, err := ctrl.groupSvc.AddGroupComment(userID, uint(articleID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, comment)
}

func (ctrl *GroupController) UpdateGroupComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID komentar tidak valid")
	}
	var req service.GroupCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	comment, err := ctrl.groupSvc.UpdateGroupComment(userID, uint(commentID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, comment)
}

func (ctrl *GroupController) DeleteGroupComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID komentar tidak valid")
	}
	if err := ctrl.groupSvc.DeleteGroupComment(userID, uint(commentID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "komentar berhasil dihapus"})
}

// ─── Reactions ────────────────────────────────────────────────────────────────

func (ctrl *GroupController) ReactToGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID artikel tidak valid")
	}
	var req service.GroupReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	result, err := ctrl.groupSvc.ReactToGroupArticle(userID, uint(articleID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}

func (ctrl *GroupController) ReactToGroupComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID komentar tidak valid")
	}
	var req service.GroupReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	result, err := ctrl.groupSvc.ReactToGroupComment(userID, uint(commentID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}
