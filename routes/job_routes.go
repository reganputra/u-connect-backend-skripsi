package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterJobRoutes(app *fiber.App, ctrl *controllers.JobController) {
	// ── Jobs (/api/jobs) ──────────────────────────────────────────────────────
	jobs := app.Group("/api/jobs", middleware.Protected())

	// GET /api/jobs/applications/mine — must be before /:id to avoid conflict
	jobs.Get("/applications/mine", middleware.RequireRole("student", "alumni"), ctrl.GetMyApplicationsHandler)
	// PUT /api/jobs/applications/:id/status
	jobs.Put("/applications/:id/status", middleware.RequireRole("alumni", "partner"), ctrl.UpdateApplicationStatusHandler)

	jobs.Get("/", ctrl.GetJobsHandler)
	jobs.Post("/", middleware.RequireRole("alumni", "partner"), ctrl.CreateJobHandler)
	jobs.Get("/:id", ctrl.GetJobByIDHandler)
	jobs.Put("/:id", middleware.RequireRole("alumni", "partner"), ctrl.UpdateJobHandler)
	jobs.Delete("/:id", middleware.RequireRole("alumni", "partner"), ctrl.DeleteJobHandler)

	// Apply for job — student and alumni
	jobs.Post("/:id/apply", middleware.RequireRole("student", "alumni"), ctrl.ApplyForJobHandler)
	// View applicants — job owner only (enforced in service)
	jobs.Get("/:id/applicants", middleware.RequireRole("alumni", "partner"), ctrl.GetApplicantsHandler)
}
