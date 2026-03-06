package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterFeedRoutes(app *fiber.App) {
	// ── Posts (/api/feed) ──────────────────────────────────────────────────────
	feed := app.Group("/api/feed", middleware.Protected())

	feed.Get("/", controllers.GetPosts)
	feed.Get("/:id", controllers.GetPostByID)
	feed.Post("/", middleware.RequireRole("alumni", "student"), controllers.CreatePost)
	feed.Put("/:id", middleware.RequireRole("alumni", "student"), controllers.UpdatePost)
	feed.Delete("/:id", middleware.RequireRole("alumni", "student"), controllers.DeletePost)

	// Add comment to a post
	feed.Post("/:id/comments", middleware.RequireRole("alumni", "student"), controllers.AddComment)

	// React & vote on a post
	feed.Post("/:id/react", controllers.ReactToPost)
	feed.Post("/:id/vote", controllers.VotePost)

	// ── Comments (/api/comments) ──────────────────────────────────────────────
	// Separate group to avoid route conflicts with /api/feed/:id
	comments := app.Group("/api/comments", middleware.Protected())

	comments.Put("/:id", middleware.RequireRole("alumni", "student"), controllers.UpdateComment)
	comments.Delete("/:id", middleware.RequireRole("alumni", "student"), controllers.DeleteComment)

	// React & vote on a comment
	comments.Post("/:id/react", controllers.ReactToComment)
	comments.Post("/:id/vote", controllers.VoteComment)
}
