package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterPortfolioRoutes(app *fiber.App, ctrl *controllers.PortfolioController) {
	portfolio := app.Group("/api/portfolio", middleware.Protected(), middleware.RequireRole("alumni", "student"))

	portfolio.Post("/", ctrl.CreatePortfolioItem)
	portfolio.Get("/", ctrl.GetPortfolioItems)
	portfolio.Put("/:id", ctrl.UpdatePortfolioItem)
	portfolio.Delete("/:id", ctrl.DeletePortfolioItem)
}
