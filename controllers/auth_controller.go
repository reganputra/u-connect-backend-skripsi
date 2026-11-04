package controllers

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/constant"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

type AuthController struct {
	authSvc service.AuthService
}

func NewAuthController(authSvc service.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req service.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	user, err := ctrl.authSvc.Register(req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, fiber.Map{
		"id":           user.ID,
		"name":         user.Name,
		"email":        user.Email,
		"role":         user.Role,
		"faculty":      user.Faculty,
		"major":        user.Major,
		"year_enroll":  user.YearEnroll,
		"company_name": user.CompanyName,
		"created_at":   user.CreatedAt,
	})
}

func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var req service.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	loginRes, err := ctrl.authSvc.Login(req)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	if loginRes.RefreshToken != "" {
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    loginRes.RefreshToken,
			Path:     "/",
			MaxAge:   constant.JWTRefreshExpiry,
			HTTPOnly: true,
			Secure:   os.Getenv("APP_ENV") != "development" || c.Protocol() == "https",
			SameSite: "Strict",
		})
	}

	c.Set("Cache-Control", "no-store")
	return utils.SuccessResponse(c, fiber.StatusOK, loginRes)
}

func (ctrl *AuthController) Refresh(c *fiber.Ctx) error {
	token := c.Cookies("refresh_token")
	loginRes, err := ctrl.authSvc.Refresh(token)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	if loginRes.RefreshToken != "" {
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    loginRes.RefreshToken,
			Path:     "/",
			MaxAge:   constant.JWTRefreshExpiry,
			HTTPOnly: true,
			Secure:   os.Getenv("APP_ENV") != "development" || c.Protocol() == "https",
			SameSite: "Strict",
		})
	}
	c.Set("Cache-Control", "no-store")
	return utils.SuccessResponse(c, fiber.StatusOK, loginRes)
}

func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}
	if err := ctrl.authSvc.Logout(userID); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") != "development" || c.Protocol() == "https",
		SameSite: "Strict",
	})
	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{"message": "berhasil logout"})
}

func (ctrl *AuthController) Me(c *fiber.Ctx) error {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "token tidak valid")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "klaim token tidak valid")
	}
	uidFloat, ok := claims["user_id"].(float64)
	if !ok || uidFloat <= 0 {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, "user_id tidak valid dalam token")
	}
	user, err := ctrl.authSvc.Me(uint(uidFloat))
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"user":    user,
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	})
}

func (ctrl *AuthController) ChangePassword(c *fiber.Ctx) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	var req service.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	if err := ctrl.authSvc.ChangePassword(userID, req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "password berhasil diperbarui",
	})
}

func (ctrl *AuthController) ForgotPassword(c *fiber.Ctx) error {
	var req service.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "isi permintaan tidak valid")
	}

	if err := ctrl.authSvc.ForgotPassword(req); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "terkunci") {
			c.Set("Retry-After", strconv.Itoa(constant.ForgotLockTTL))
			// Also set rate-limit style header for 30m
			c.Set("X-Retry-After", time.Now().Add(time.Duration(constant.ForgotLockTTL)*time.Second).Format(time.RFC3339))
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, msg)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "password berhasil direset, silakan login dengan password baru",
	})
}

func (ctrl *AuthController) UnlockUserReset(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "id tidak valid")
	}

	if err := ctrl.authSvc.UnlockUserReset(uint(id)); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, fiber.Map{
		"message": "akun berhasil dibuka kuncinya",
	})
}
