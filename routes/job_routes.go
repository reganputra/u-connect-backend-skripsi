package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterJobRoutes(app *fiber.App) {
	// ── Jobs (/api/jobs) ──────────────────────────────────────────────────────
	jobs := app.Group("/api/jobs", middleware.Protected())

	// GET /api/jobs/applications/mine — must be before /:id to avoid conflict
	jobs.Get("/applications/mine", middleware.RequireRole("student", "alumni"), controllers.GetMyApplicationsHandler)
	// PUT /api/jobs/applications/:id/status
	jobs.Put("/applications/:id/status", middleware.RequireRole("alumni", "partner"), controllers.UpdateApplicationStatusHandler)

	jobs.Get("/", controllers.GetJobsHandler)
	jobs.Post("/", middleware.RequireRole("alumni", "partner"), controllers.CreateJobHandler)
	jobs.Get("/:id", controllers.GetJobByIDHandler)
	jobs.Put("/:id", middleware.RequireRole("alumni", "partner"), controllers.UpdateJobHandler)
	jobs.Delete("/:id", middleware.RequireRole("alumni", "partner"), controllers.DeleteJobHandler)

	// Apply for job — student and alumni
	jobs.Post("/:id/apply", middleware.RequireRole("student", "alumni"), controllers.ApplyForJobHandler)
	// View applicants — job owner only (enforced in service)
	jobs.Get("/:id/applicants", middleware.RequireRole("alumni", "partner"), controllers.GetApplicantsHandler)
}
