package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterGroupRoutes(app *fiber.App, ctrl *controllers.GroupController) {
	// ── Groups (/api/groups) ───────────────────────────────────────────────────
	groups := app.Group("/api/groups", middleware.Protected())

	groups.Get("/", ctrl.GetGroups)
	groups.Get("/joined", middleware.RequireRole("alumni", "student"), ctrl.GetJoinedGroups)
	groups.Post("/", middleware.RequireRole("alumni", "student"), ctrl.CreateGroup)
	groups.Get("/:id", ctrl.GetGroupByID)
	groups.Put("/:id", middleware.RequireRole("alumni", "student"), ctrl.UpdateGroup)
	groups.Delete("/:id", middleware.RequireRole("alumni", "student"), ctrl.DeleteGroup)

	// Membership
	groups.Post("/:id/join", middleware.RequireRole("alumni", "student"), ctrl.JoinGroup)
	groups.Delete("/:id/leave", middleware.RequireRole("alumni", "student"), ctrl.LeaveGroup)
	groups.Get("/:id/members", ctrl.GetGroupMembers)
	groups.Delete("/:id/members/:userID", middleware.RequireRole("alumni", "student"), ctrl.KickMember)

	// Articles inside a group (form-data)
	groups.Post("/:id/articles", middleware.RequireRole("alumni", "student"), ctrl.CreateGroupArticle)

	// ── Articles (/api/groups/articles) ───────────────────────────────────────
	// Separate group to avoid conflicts with /:id param
	articles := app.Group("/api/groups/articles", middleware.Protected())

	articles.Get("/:id", ctrl.GetGroupArticleDetail)
	articles.Put("/:id", middleware.RequireRole("alumni", "student"), ctrl.UpdateGroupArticle)
	articles.Delete("/:id", middleware.RequireRole("alumni", "student"), ctrl.DeleteGroupArticle)
	articles.Post("/:id/comments", middleware.RequireRole("alumni", "student"), ctrl.AddGroupComment)
	articles.Post("/:id/react", middleware.RequireRole("alumni", "student"), ctrl.ReactToGroupArticle)

	// ── Comments (/api/groups/comments) ───────────────────────────────────────
	comments := app.Group("/api/groups/comments", middleware.Protected())

	comments.Put("/:id", middleware.RequireRole("alumni", "student"), ctrl.UpdateGroupComment)
	comments.Delete("/:id", middleware.RequireRole("alumni", "student"), ctrl.DeleteGroupComment)
	comments.Post("/:id/react", middleware.RequireRole("alumni", "student"), ctrl.ReactToGroupComment)
}
