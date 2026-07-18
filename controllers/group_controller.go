package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/dto"
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
	page, limit := utils.ParsePagination(c, 20)

	groups, total, err := ctrl.groupSvc.GetGroups(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data grup")
	}
	return utils.PaginatedResponse(c, fiber.StatusOK, groups, total, page, limit)
}

func (ctrl *GroupController) GetOwnedGroups(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	page, limit := utils.ParsePagination(c, 20)

	groups, total, err := ctrl.groupSvc.GetOwnedGroups(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data grup milik pengguna")
	}

	return utils.PaginatedResponse(c, fiber.StatusOK, groups, total, page, limit)
}

func (ctrl *GroupController) GetGroupByID(c *fiber.Ctx) error {
	groupID, ok := utils.MustParseIDParam(c, "id", "grup")
	if !ok {
		return nil
	}
	articlePage, _ := strconv.Atoi(c.Query("article_page", "1"))
	articleLimit, _ := strconv.Atoi(c.Query("article_limit", "10"))
	group, err := ctrl.groupSvc.GetGroupByID(groupID, articlePage, articleLimit)
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
	groupID, ok := utils.MustParseIDParam(c, "id", "grup")
	if !ok {
		return nil
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
	group, err := ctrl.groupSvc.UpdateGroup(userID, groupID, req)
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
	groupID, ok := utils.MustParseIDParam(c, "id", "grup")
	if !ok {
		return nil
	}
	if err := ctrl.groupSvc.DeleteGroup(userID, groupID); err != nil {
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
	groupID, ok := utils.MustParseIDParam(c, "id", "grup")
	if !ok {
		return nil
	}
	if err := ctrl.groupSvc.JoinGroup(userID, groupID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil bergabung dengan grup"})
}

func (ctrl *GroupController) LeaveGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, ok := utils.MustParseIDParam(c, "id", "grup")
	if !ok {
		return nil
	}
	if err := ctrl.groupSvc.LeaveGroup(userID, groupID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil meninggalkan grup"})
}

func (ctrl *GroupController) GetGroupMembers(c *fiber.Ctx) error {
	groupID, ok := utils.MustParseIDParam(c, "id", "grup")
	if !ok {
		return nil
	}
	page, limit := utils.ParsePagination(c, 20)

	members, total, err := ctrl.groupSvc.GetGroupMembers(groupID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data anggota")
	}
	return utils.PaginatedResponse(c, fiber.StatusOK, members, total, page, limit)
}

func (ctrl *GroupController) GetJoinedGroups(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	page, limit := utils.ParsePagination(c, 20)

	groups, total, err := ctrl.groupSvc.GetJoinedGroups(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data grup yang diikuti")
	}
	return utils.PaginatedResponse(c, fiber.StatusOK, groups, total, page, limit)
}

func (ctrl *GroupController) KickMember(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, ok := utils.MustParseIDParam(c, "id", "grup")
	if !ok {
		return nil
	}
	targetID, ok := utils.MustParseIDParam(c, "userID", "pengguna")
	if !ok {
		return nil
	}
	var body dto.KickMemberRequest
	_ = c.BodyParser(&body)
	if err := ctrl.groupSvc.KickMember(userID, groupID, targetID, body.Reason); err != nil {
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
	groupID, ok := utils.MustParseIDParam(c, "id", "grup")
	if !ok {
		return nil
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
	article, err := ctrl.groupSvc.CreateGroupArticle(userID, groupID, req)
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
	articleID, ok := utils.MustParseIDParam(c, "id", "artikel")
	if !ok {
		return nil
	}
	detail, err := ctrl.groupSvc.GetGroupArticleDetail(userID, articleID)
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
	articleID, ok := utils.MustParseIDParam(c, "id", "artikel")
	if !ok {
		return nil
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
	article, err := ctrl.groupSvc.UpdateGroupArticle(userID, articleID, req)
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
	articleID, ok := utils.MustParseIDParam(c, "id", "artikel")
	if !ok {
		return nil
	}
	if err := ctrl.groupSvc.DeleteGroupArticle(userID, articleID); err != nil {
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
	articleID, ok := utils.MustParseIDParam(c, "id", "artikel")
	if !ok {
		return nil
	}
	var req service.GroupCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	comment, err := ctrl.groupSvc.AddGroupComment(userID, articleID, req)
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
	commentID, ok := utils.MustParseIDParam(c, "id", "komentar")
	if !ok {
		return nil
	}
	var req service.GroupCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	comment, err := ctrl.groupSvc.UpdateGroupComment(userID, commentID, req)
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
	commentID, ok := utils.MustParseIDParam(c, "id", "komentar")
	if !ok {
		return nil
	}
	if err := ctrl.groupSvc.DeleteGroupComment(userID, commentID); err != nil {
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
	articleID, ok := utils.MustParseIDParam(c, "id", "artikel")
	if !ok {
		return nil
	}
	var req service.GroupReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	result, err := ctrl.groupSvc.ReactToGroupArticle(userID, articleID, req)
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
	commentID, ok := utils.MustParseIDParam(c, "id", "komentar")
	if !ok {
		return nil
	}
	var req service.GroupReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}
	result, err := ctrl.groupSvc.ReactToGroupComment(userID, commentID, req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}
