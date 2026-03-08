package controllers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

// ─── Group CRUD ───────────────────────────────────────────────────────────────

func CreateGroup(c *fiber.Ctx) error {
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
	group, err := service.CreateGroup(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, group)
}

func GetGroups(c *fiber.Ctx) error {
	groups, err := service.GetGroups()
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data grup")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, groups)
}

func GetGroupByID(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	group, err := service.GetGroupByID(uint(groupID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, group)
}

func UpdateGroup(c *fiber.Ctx) error {
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
	group, err := service.UpdateGroup(userID, uint(groupID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya pemilik grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, group)
}

func DeleteGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	if err := service.DeleteGroup(userID, uint(groupID)); err != nil {
		if err.Error() == "akses ditolak: hanya pemilik grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "grup berhasil dihapus"})
}

// ─── Membership ───────────────────────────────────────────────────────────────

func JoinGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	if err := service.JoinGroup(userID, uint(groupID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil bergabung dengan grup"})
}

func LeaveGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	if err := service.LeaveGroup(userID, uint(groupID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil meninggalkan grup"})
}

func GetGroupMembers(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	members, err := service.GetGroupMembers(uint(groupID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data anggota")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, members)
}

func GetJoinedGroups(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groups, err := service.GetJoinedGroups(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data grup yang diikuti")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, groups)
}

func KickMember(c *fiber.Ctx) error {
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
	if err := service.KickMember(userID, uint(groupID), uint(targetID)); err != nil {
		if err.Error() == "akses ditolak: hanya pemilik grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "anggota berhasil dikeluarkan"})
}

// ─── Articles ─────────────────────────────────────────────────────────────────

func CreateGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID grup tidak valid")
	}
	mediaURL, err := uploadFileIfPresent(c, "media", "alumni-platform/groups/articles")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	req := service.GroupArticleRequest{
		Title:    c.FormValue("title"),
		Content:  c.FormValue("content"),
		MediaURL: parseOptionalString(mediaURL),
	}
	article, err := service.CreateGroupArticle(userID, uint(groupID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, article)
}

func GetGroupArticleDetail(c *fiber.Ctx) error {
	userID, _ := getUserIDFromToken(c)
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID artikel tidak valid")
	}
	detail, err := service.GetGroupArticleDetail(userID, uint(articleID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, detail)
}

func UpdateGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID artikel tidak valid")
	}
	mediaURL, err := uploadFileIfPresent(c, "media", "alumni-platform/groups/articles")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	req := service.GroupArticleRequest{
		Title:    c.FormValue("title"),
		Content:  c.FormValue("content"),
		MediaURL: parseOptionalString(mediaURL),
	}
	article, err := service.UpdateGroupArticle(userID, uint(articleID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, article)
}

func DeleteGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID artikel tidak valid")
	}
	if err := service.DeleteGroupArticle(userID, uint(articleID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "artikel berhasil dihapus"})
}

// ─── Comments ─────────────────────────────────────────────────────────────────

func AddGroupComment(c *fiber.Ctx) error {
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
	comment, err := service.AddGroupComment(userID, uint(articleID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, comment)
}

func UpdateGroupComment(c *fiber.Ctx) error {
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
	comment, err := service.UpdateGroupComment(userID, uint(commentID), req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, comment)
}

func DeleteGroupComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	commentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "ID komentar tidak valid")
	}
	if err := service.DeleteGroupComment(userID, uint(commentID)); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "komentar berhasil dihapus"})
}

// ─── Reactions ────────────────────────────────────────────────────────────────

func ReactToGroupArticle(c *fiber.Ctx) error {
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
	result, err := service.ReactToGroupArticle(userID, uint(articleID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}

func ReactToGroupComment(c *fiber.Ctx) error {
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
	result, err := service.ReactToGroupComment(userID, uint(commentID), req)
	if err != nil {
		if err.Error() == "akses ditolak: hanya anggota grup" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}

// keep config import used
var _ = config.DB
