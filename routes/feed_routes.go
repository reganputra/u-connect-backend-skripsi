package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterFeedRoutes(app *fiber.App, ctrl *controllers.FeedController) {
	// ── Posts (/api/feed) ──────────────────────────────────────────────────────
	feed := app.Group("/api/feed", middleware.Protected())

	feed.Get("/", ctrl.GetPosts)
	feed.Get("/:id", ctrl.GetPostByID)
	feed.Post("/", middleware.RequireRole("alumni", "student"), ctrl.CreatePost)
	feed.Put("/:id", middleware.RequireRole("alumni", "student"), ctrl.UpdatePost)
	feed.Delete("/:id", middleware.RequireRole("alumni", "student"), ctrl.DeletePost)

	// Add comment to a post
	feed.Post("/:id/comments", middleware.RequireRole("alumni", "student"), ctrl.AddComment)

	// React & vote on a post
	feed.Post("/:id/react", ctrl.ReactToPost)
	feed.Post("/:id/vote", ctrl.VotePost)

	// ── Comments (/api/comments) ──────────────────────────────────────────────
	// Separate group to avoid route conflicts with /api/feed/:id
	comments := app.Group("/api/comments", middleware.Protected())

	comments.Put("/:id", middleware.RequireRole("alumni", "student"), ctrl.UpdateComment)
	comments.Delete("/:id", middleware.RequireRole("alumni", "student"), ctrl.DeleteComment)

	// React & vote on a comment
	comments.Post("/:id/react", ctrl.ReactToComment)
	comments.Post("/:id/vote", ctrl.VoteComment)
}
