package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterPortfolioRoutes(app *fiber.App) {
	portfolio := app.Group("/api/portfolio",
		middleware.Protected(),
		middleware.RequireRole("alumni", "student"),
	)

	portfolio.Post("/", controllers.CreatePortfolioItem)
	portfolio.Get("/", controllers.GetPortfolioItems)
	portfolio.Put("/:id", controllers.UpdatePortfolioItem)
	portfolio.Delete("/:id", controllers.DeletePortfolioItem)
}
