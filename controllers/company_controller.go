package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/dto"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type CompanyController struct {
	companySvc service.CompanyService
}

func NewCompanyController(companySvc service.CompanyService) *CompanyController {
	return &CompanyController{companySvc: companySvc}
}

func (ctrl *CompanyController) CreateOrJoinCompanyProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	var req service.CompanyProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}

	profile, created, err := ctrl.companySvc.CreateOrJoinCompanyProfile(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	status := fiber.StatusOK
	if created {
		status = fiber.StatusCreated
	}
	return utils.SuccessResponse(c, status, profile)
}

func (ctrl *CompanyController) GetCompanyProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	profile, err := ctrl.companySvc.GetCompanyProfile(userID)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

func (ctrl *CompanyController) UpdateCompanyProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	var req service.CompanyProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}

	profile, err := ctrl.companySvc.UpdateCompanyProfile(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, profile)
}

func (ctrl *CompanyController) ChangeCompanyAffiliation(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	var body dto.ChangeCompanyAffiliationRequest
	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, dto.MsgInvalidRequest)
	}

	profile, joined, err := ctrl.companySvc.ChangeCompanyAffiliation(userID, body.CompanyName)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"joined_existing": joined,
		"company":         profile,
	})
}

func (ctrl *CompanyController) DeleteCompanyProfile(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	if err := ctrl.companySvc.DeleteCompanyProfile(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "profil perusahaan berhasil dihapus"})
}
