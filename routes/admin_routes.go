package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterAdminRoutes(app *fiber.App, ctrl *controllers.AdminController, authCtrl *controllers.AuthController, evalCtrl *controllers.EvaluationController) {
	// Public: category list (any authenticated user can read categories)
	app.Get("/api/categories", middleware.Protected(), ctrl.GetCategories)

	// All admin routes: require JWT + admin role
	admin := app.Group("/api/admin", middleware.Protected(), middleware.RequireRole("admin"))

	// Dashboard
	admin.Get("/dashboard", ctrl.GetDashboard)

	// User management
	admin.Get("/users", ctrl.GetUsers)
	admin.Get("/users/:id", ctrl.GetUserByID)
	admin.Patch("/users/:id/status", ctrl.SetUserStatus)
	admin.Patch("/users/:id/role", ctrl.SetUserRole)
	admin.Patch("/users/:id/unlock-reset", authCtrl.UnlockUserReset)
	admin.Patch("/users/:id/profile", ctrl.PatchUserProfile)
	admin.Post("/users/:id/experience", ctrl.AddUserExperience)
	admin.Put("/users/:id/experience/:expId", ctrl.UpdateUserExperience)
	admin.Delete("/users/:id/experience/:expId", ctrl.DeleteUserExperience)

	// Report moderation
	admin.Get("/reports", ctrl.GetReports)
	admin.Get("/reports/:id", ctrl.GetReportByID)
	admin.Patch("/reports/:id/resolve", ctrl.ResolveReport)
	admin.Patch("/reports/:id/reject", ctrl.RejectReport)

	// Direct content deletion
	admin.Delete("/posts/:id", ctrl.DeletePost)
	admin.Delete("/groups/:id", ctrl.DeleteGroup)
	admin.Delete("/events/:id", ctrl.DeleteEvent)
	admin.Delete("/jobs/:id", ctrl.DeleteJob)

	// Category management (admin only for write)
	admin.Post("/categories", ctrl.CreateCategory)
	admin.Put("/categories/:id", ctrl.UpdateCategory)
	admin.Delete("/categories/:id", ctrl.DeleteCategory)

	// CBF Evaluation (MAP metric) — admin-only, for academic/thesis evaluation
	admin.Get("/evaluation/cbf", evalCtrl.EvaluateCBF)
	admin.Get("/evaluation/cbf-no-lemma", evalCtrl.EvaluateCBFWithoutLemmatizer)

	// CBF Evaluation (MRR metric) — admin-only, untuk evaluasi skripsi
	admin.Get("/evaluation/cbf-mrr", evalCtrl.EvaluateCBFMRR)
	admin.Get("/evaluation/cbf-mrr-no-lemma", evalCtrl.EvaluateCBFMRRWithoutLemmatizer)
}
