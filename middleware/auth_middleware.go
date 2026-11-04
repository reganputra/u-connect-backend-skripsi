package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/constant"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/utils"
)

// AuthUserRepositoryKey is request-scoped: inject before registering protected routes.
const AuthUserRepositoryKey = "auth_user_repository"

// AuthRepository binds an explicit application dependency without global state.
func AuthRepository(repo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error { c.Locals(AuthUserRepositoryKey, repo); return c.Next() }
}

// Protected validates an access token and checks current account state on every request.
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		parts := strings.Fields(c.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return unauthorized(c)
		}
		claims, err := utils.ParseJWT(parts[1])
		if err != nil {
			return unauthorized(c)
		}
		repo, ok := c.Locals(AuthUserRepositoryKey).(repository.UserRepository)
		if !ok || repo == nil {
			return unauthorized(c)
		}
		if !validAccount(repo, claims) {
			return unauthorized(c)
		}
		c.Locals("user", &jwt.Token{Claims: claims, Valid: true, Method: jwt.SigningMethodHS256})
		return c.Next()
	}
}

func unauthorized(c *fiber.Ctx) error {
	return utils.ErrorResponse(c, fiber.StatusUnauthorized, "token tidak ada atau sesi tidak valid")
}

func validAccount(repo repository.UserRepository, claims jwt.MapClaims) bool {
	if repo == nil {
		return false
	}
	id, version, err := utils.TokenIdentity(claims)
	if err != nil || claims["token_type"] != "access" {
		return false
	}
	user, err := repo.FindUserByID(id)
	return err == nil && user != nil && user.IsActive && user.TokenVersion == version && user.Role == claims["role"]
}

// TokenVersionGuard remains for callers that explicitly supply a repository.
// Protected already performs this check; this guard never bypasses missing state.
func TokenVersionGuard(repo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, ok := c.Locals("user").(*jwt.Token)
		if !ok || token == nil || !token.Valid {
			return unauthorized(c)
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !validAccount(repo, claims) {
			return unauthorized(c)
		}
		return c.Next()
	}
}

// RequireRole mengembalikan middleware yang hanya mengizinkan peran yang ditentukan.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*jwt.Token)
		if !ok {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "token tidak ada atau tidak valid")
		}

		claims, ok := user.Claims.(jwt.MapClaims)
		if !ok {
			return utils.ErrorResponse(c, fiber.StatusUnauthorized, "klaim token tidak valid")
		}

		role, _ := claims["role"].(string)
		if role == constant.RoleAdmin {
			return c.Next()
		}
		for _, allowed := range roles {
			if role == allowed {
				return c.Next()
			}
		}
		return utils.ErrorResponse(c, fiber.StatusForbidden, "akses ditolak: peran tidak mencukupi")
	}
}
