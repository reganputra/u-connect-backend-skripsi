package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterCompanyRoutes(app *fiber.App) {
	company := app.Group("/api/company",
		middleware.Protected(),
		middleware.RequireRole("partner"),
	)

	company.Post("/", controllers.CreateOrJoinCompanyProfile)
	company.Get("/", controllers.GetCompanyProfile)
	company.Put("/", controllers.UpdateCompanyProfile)
	company.Delete("/", controllers.DeleteCompanyProfile)
}
