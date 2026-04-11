package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterCompanyRoutes(app *fiber.App, ctrl *controllers.CompanyController) {
	company := app.Group("/api/company", middleware.Protected(), middleware.RequireRole("partner"))

	company.Post("/", ctrl.CreateOrJoinCompanyProfile)
	company.Get("/", ctrl.GetCompanyProfile)
	company.Put("/", ctrl.UpdateCompanyProfile)
	company.Patch("/affiliation", ctrl.ChangeCompanyAffiliation)
	company.Delete("/", ctrl.DeleteCompanyProfile)
}
