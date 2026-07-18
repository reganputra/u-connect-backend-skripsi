package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type FollowController struct {
	followSvc service.FollowService
}

func NewFollowController(followSvc service.FollowService) *FollowController {
	return &FollowController{followSvc: followSvc}
}

func (ctrl *FollowController) Follow(c *fiber.Ctx) error {
	callerID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	role, err := getUserRoleFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	targetID, ok := utils.MustParseIDParam(c, "id", "pengguna")
	if !ok {
		return nil
	}

	if err := ctrl.followSvc.Follow(callerID, uint(targetID), role); err != nil {
		switch err.Error() {
		case "anda sudah mengikuti pengguna ini":
			return utils.ErrorResponse(c, fiber.StatusConflict, err.Error())
		case "tidak dapat mengikuti diri sendiri":
			return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
		case "hanya student dan alumni yang dapat mengikuti pengguna":
			return utils.ErrorResponse(c, fiber.StatusForbidden, err.Error())
		default:
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
		}
	}
	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{"message": "berhasil mengikuti pengguna"})
}

func (ctrl *FollowController) Unfollow(c *fiber.Ctx) error {
	callerID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	targetID, ok := utils.MustParseIDParam(c, "id", "pengguna")
	if !ok {
		return nil
	}

	if err := ctrl.followSvc.Unfollow(callerID, uint(targetID)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil berhenti mengikuti pengguna"})
}

func (ctrl *FollowController) GetFollowers(c *fiber.Ctx) error {

	targetID, ok := utils.MustParseIDParam(c, "id", "pengguna")
	if !ok {
		return nil
	}

	users, err := ctrl.followSvc.GetFollowers(uint(targetID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, users)
}

func (ctrl *FollowController) GetFollowing(c *fiber.Ctx) error {

	targetID, ok := utils.MustParseIDParam(c, "id", "pengguna")
	if !ok {
		return nil
	}

	users, err := ctrl.followSvc.GetFollowing(uint(targetID))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, users)
}
