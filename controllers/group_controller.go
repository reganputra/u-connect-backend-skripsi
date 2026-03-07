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
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to fetch groups")
	}
	return utils.SuccessResponse(c, fiber.StatusOK, groups)
}

func GetGroupByID(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid group id")
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid group id")
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
		if err.Error() == "access denied: owner only" {
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid group id")
	}
	if err := service.DeleteGroup(userID, uint(groupID)); err != nil {
		if err.Error() == "access denied: owner only" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "group deleted successfully"})
}

// ─── Membership ───────────────────────────────────────────────────────────────

func JoinGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid group id")
	}
	if err := service.JoinGroup(userID, uint(groupID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "joined group successfully"})
}

func LeaveGroup(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid group id")
	}
	if err := service.LeaveGroup(userID, uint(groupID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "left group successfully"})
}

func GetGroupMembers(c *fiber.Ctx) error {
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid group id")
	}
	members, err := service.GetGroupMembers(uint(groupID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to fetch members")
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
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "failed to fetch joined groups")
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid group id")
	}
	targetID, err := strconv.ParseUint(c.Params("userID"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid user id")
	}
	if err := service.KickMember(userID, uint(groupID), uint(targetID)); err != nil {
		if err.Error() == "access denied: owner only" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "member removed successfully"})
}

// ─── Articles ─────────────────────────────────────────────────────────────────

func CreateGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	groupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid group id")
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
		if err.Error() == "access denied: members only" {
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid article id")
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid article id")
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
		if err.Error() == "access denied" {
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid article id")
	}
	if err := service.DeleteGroupArticle(userID, uint(articleID)); err != nil {
		if err.Error() == "access denied" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "article deleted successfully"})
}

// ─── Comments ─────────────────────────────────────────────────────────────────

func AddGroupComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid article id")
	}
	var req service.GroupCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	comment, err := service.AddGroupComment(userID, uint(articleID), req)
	if err != nil {
		if err.Error() == "access denied: members only" {
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid comment id")
	}
	var req service.GroupCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	comment, err := service.UpdateGroupComment(userID, uint(commentID), req)
	if err != nil {
		if err.Error() == "access denied" {
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid comment id")
	}
	if err := service.DeleteGroupComment(userID, uint(commentID)); err != nil {
		if err.Error() == "access denied" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "comment deleted successfully"})
}

// ─── Reactions ────────────────────────────────────────────────────────────────

func ReactToGroupArticle(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	articleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid article id")
	}
	var req service.GroupReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	result, err := service.ReactToGroupArticle(userID, uint(articleID), req)
	if err != nil {
		if err.Error() == "access denied: members only" {
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
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid comment id")
	}
	var req service.GroupReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "invalid request body")
	}
	result, err := service.ReactToGroupComment(userID, uint(commentID), req)
	if err != nil {
		if err.Error() == "access denied: members only" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}

// keep config import used
var _ = config.DB
