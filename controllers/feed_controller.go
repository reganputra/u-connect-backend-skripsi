package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/dto"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type FeedController struct {
	feedSvc service.FeedService
}

func NewFeedController(feedSvc service.FeedService) *FeedController {
	return &FeedController{feedSvc: feedSvc}
}

// ─── Posts ────────────────────────────────────────────────────────────────────

func (ctrl *FeedController) CreatePost(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	imageURLs, err := uploadFilesIfPresent(c, "images", "alumni-platform/feed")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	// Also support legacy single "image" field for backward compatibility
	if len(imageURLs) == 0 {
		imageURL, err := uploadFileIfPresent(c, "image", "alumni-platform/feed")
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		if imageURL != "" {
			imageURLs = []string{imageURL}
		}
	}

	req := service.PostRequest{
		Category:  parseOptionalString(c.FormValue("category")),
		Title:     c.FormValue("title"),
		Content:   c.FormValue("content"),
		ImageURLs: imageURLs,
	}

	post, err := ctrl.feedSvc.CreatePost(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, post)
}

func (ctrl *FeedController) GetPosts(c *fiber.Ctx) error {
	page, limit := utils.ParsePagination(c, 10)

	posts, total, err := ctrl.feedSvc.GetPosts(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "gagal mengambil data postingan")
	}
	return utils.PaginatedResponse(c, fiber.StatusOK, posts, total, page, limit)
}

func (ctrl *FeedController) GetPostByID(c *fiber.Ctx) error {

	postID, ok := utils.MustParseIDParam(c, "id", "postingan")
	if !ok {
		return nil
	}
	post, err := ctrl.feedSvc.GetPostByID(postID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, post)
}

func (ctrl *FeedController) UpdatePost(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	postID, ok := utils.MustParseIDParam(c, "id", "postingan")
	if !ok {
		return nil
	}

	imageURLs, err := uploadFilesIfPresent(c, "images", "alumni-platform/feed")
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	// Also support legacy single "image" field for backward compatibility
	if len(imageURLs) == 0 {
		imageURL, err := uploadFileIfPresent(c, "image", "alumni-platform/feed")
		if err != nil {
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
		if imageURL != "" {
			imageURLs = []string{imageURL}
		}
	}

	req := service.PostRequest{
		Category:  parseOptionalString(c.FormValue("category")),
		Title:     c.FormValue("title"),
		Content:   c.FormValue("content"),
		ImageURLs: imageURLs,
	}

	post, err := ctrl.feedSvc.UpdatePost(userID, postID, req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, post)
}

func (ctrl *FeedController) DeletePost(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	postID, ok := utils.MustParseIDParam(c, "id", "postingan")
	if !ok {
		return nil
	}
	if err := ctrl.feedSvc.DeletePost(userID, postID); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "postingan berhasil dihapus"})
}

// ─── Comments ─────────────────────────────────────────────────────────────────

func (ctrl *FeedController) AddComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	postID, ok := utils.MustParseIDParam(c, "id", "postingan")
	if !ok {
		return nil
	}
	var req service.CommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}
	comment, err := ctrl.feedSvc.AddComment(userID, postID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, comment)
}

func (ctrl *FeedController) UpdateComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	commentID, ok := utils.MustParseIDParam(c, "id", "komentar")
	if !ok {
		return nil
	}
	var req service.CommentRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}
	comment, err := ctrl.feedSvc.UpdateComment(userID, commentID, req)
	if err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, comment)
}

func (ctrl *FeedController) DeleteComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	commentID, ok := utils.MustParseIDParam(c, "id", "komentar")
	if !ok {
		return nil
	}
	if err := ctrl.feedSvc.DeleteComment(userID, commentID); err != nil {
		if err.Error() == "akses ditolak" {
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "komentar berhasil dihapus"})
}

// ─── Reactions ────────────────────────────────────────────────────────────────

func (ctrl *FeedController) ReactToPost(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	postID, ok := utils.MustParseIDParam(c, "id", "postingan")
	if !ok {
		return nil
	}
	var req service.ReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}
	result, err := ctrl.feedSvc.ReactToPost(userID, postID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}

func (ctrl *FeedController) ReactToComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	commentID, ok := utils.MustParseIDParam(c, "id", "komentar")
	if !ok {
		return nil
	}
	var req service.ReactionRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}
	result, err := ctrl.feedSvc.ReactToComment(userID, commentID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}

// ─── Votes ────────────────────────────────────────────────────────────────────

func (ctrl *FeedController) VotePost(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	postID, ok := utils.MustParseIDParam(c, "id", "postingan")
	if !ok {
		return nil
	}
	var req service.VoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}
	result, err := ctrl.feedSvc.VotePost(userID, postID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}

func (ctrl *FeedController) VoteComment(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	commentID, ok := utils.MustParseIDParam(c, "id", "komentar")
	if !ok {
		return nil
	}
	var req service.VoteRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}
	result, err := ctrl.feedSvc.VoteComment(userID, commentID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"action": result})
}
