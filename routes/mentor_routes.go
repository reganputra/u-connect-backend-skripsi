package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterMentorRoutes(app *fiber.App, ctrl *controllers.MentorController) {
	// ── Public mentor browsing (student only) ──────────────────────────────────
	// /api/mentors/recommend MUST be registered before /api/mentors/:id to avoid
	// Fiber treating "recommend" as an :id parameter value.
	mentors := app.Group("/api/mentors", middleware.Protected())
	mentors.Get("/recommend", middleware.RequireRole("student"), ctrl.GetRecommendations)
	mentors.Get("/", middleware.RequireRole("student"), ctrl.GetMentors)
	mentors.Get("/:id", middleware.RequireRole("student"), ctrl.GetMentorDetail)
	mentors.Post("/:id/request", middleware.RequireRole("student"), ctrl.RequestMentoring)

	// ── Student dashboard ──────────────────────────────────────────────────────
	student := app.Group("/api/student", middleware.Protected(), middleware.RequireRole("student"))
	student.Get("/mentors", ctrl.GetMyMentors)
	student.Get("/requests", ctrl.GetSentRequests)
	student.Delete("/requests/:id", ctrl.WithdrawRequest)
	student.Post("/sessions", ctrl.CreateSessionAsStudent)
	student.Get("/sessions", ctrl.GetStudentSessions)

	// ── Mentor management (alumni only) ───────────────────────────────────────
	mentor := app.Group("/api/mentor", middleware.Protected(), middleware.RequireRole("alumni"))
	mentor.Post("/register", ctrl.RegisterAsMentor)
	mentor.Get("/profile", ctrl.GetMyMentorProfile)
	mentor.Put("/profile", ctrl.UpdateMentorProfile)
	mentor.Delete("/unregister", ctrl.UnregisterAsMentor)
	mentor.Get("/requests", ctrl.GetIncomingRequests)
	mentor.Patch("/requests/:id/approve", ctrl.ApproveRequest)
	mentor.Patch("/requests/:id/reject", ctrl.RejectRequest)
	mentor.Get("/mentees", ctrl.GetMyMentees)
	mentor.Post("/sessions", ctrl.CreateSession)
	mentor.Get("/sessions", ctrl.GetMentorSessions)
	mentor.Patch("/sessions/:id", ctrl.UpdateSession)
}
