package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func SetupFollowRoutes(app *fiber.App, ctrl *controllers.FollowController) {
	// All follow endpoints require authentication.
	// Role restriction is enforced inside FollowService.Follow (student/alumni only).
	users := app.Group("/api/users", middleware.Protected())

	users.Post("/:id/follow", ctrl.Follow)
	users.Delete("/:id/follow", ctrl.Unfollow)
	users.Get("/:id/followers", ctrl.GetFollowers)
	users.Get("/:id/following", ctrl.GetFollowing)
}
