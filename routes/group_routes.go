package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterGroupRoutes(app *fiber.App) {
	// ── Groups (/api/groups) ───────────────────────────────────────────────────
	groups := app.Group("/api/groups", middleware.Protected())

	groups.Get("/", controllers.GetGroups)
	groups.Get("/joined", middleware.RequireRole("alumni", "student"), controllers.GetJoinedGroups)
	groups.Post("/", middleware.RequireRole("alumni", "student"), controllers.CreateGroup)
	groups.Get("/:id", controllers.GetGroupByID)
	groups.Put("/:id", middleware.RequireRole("alumni", "student"), controllers.UpdateGroup)
	groups.Delete("/:id", middleware.RequireRole("alumni", "student"), controllers.DeleteGroup)

	// Membership
	groups.Post("/:id/join", middleware.RequireRole("alumni", "student"), controllers.JoinGroup)
	groups.Delete("/:id/leave", middleware.RequireRole("alumni", "student"), controllers.LeaveGroup)
	groups.Get("/:id/members", controllers.GetGroupMembers)
	groups.Delete("/:id/members/:userID", middleware.RequireRole("alumni", "student"), controllers.KickMember)

	// Articles inside a group (form-data)
	groups.Post("/:id/articles", middleware.RequireRole("alumni", "student"), controllers.CreateGroupArticle)

	// ── Articles (/api/groups/articles) ───────────────────────────────────────
	// Separate group to avoid conflicts with /:id param
	articles := app.Group("/api/groups/articles", middleware.Protected())

	articles.Get("/:id", controllers.GetGroupArticleDetail)
	articles.Put("/:id", middleware.RequireRole("alumni", "student"), controllers.UpdateGroupArticle)
	articles.Delete("/:id", middleware.RequireRole("alumni", "student"), controllers.DeleteGroupArticle)
	articles.Post("/:id/comments", middleware.RequireRole("alumni", "student"), controllers.AddGroupComment)
	articles.Post("/:id/react", middleware.RequireRole("alumni", "student"), controllers.ReactToGroupArticle)

	// ── Comments (/api/groups/comments) ───────────────────────────────────────
	comments := app.Group("/api/groups/comments", middleware.Protected())

	comments.Put("/:id", middleware.RequireRole("alumni", "student"), controllers.UpdateGroupComment)
	comments.Delete("/:id", middleware.RequireRole("alumni", "student"), controllers.DeleteGroupComment)
	comments.Post("/:id/react", middleware.RequireRole("alumni", "student"), controllers.ReactToGroupComment)
}
