package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
)

func RegisterEventRoutes(app *fiber.App) {
	// ── Events (/api/events) ──────────────────────────────────────────────────
	events := app.Group("/api/events", middleware.Protected())

	events.Get("/", controllers.GetEvents)
	events.Post("/", middleware.RequireRole("alumni", "student"), controllers.CreateEvent)
	events.Get("/:id", controllers.GetEventByID)
	events.Put("/:id", middleware.RequireRole("alumni", "student"), controllers.UpdateEvent)
	events.Delete("/:id", middleware.RequireRole("alumni", "student"), controllers.DeleteEvent)

	// Registration
	events.Post("/:id/register", middleware.RequireRole("alumni", "student"), controllers.RegisterForEvent)
	events.Delete("/:id/register", middleware.RequireRole("alumni", "student"), controllers.CancelRegistration)
	events.Get("/:id/participants", controllers.GetParticipants)

	// Agenda (add to a specific event)
	events.Post("/:id/agenda", middleware.RequireRole("alumni", "student"), controllers.AddAgenda)

	// ── Agenda (/api/events/agenda) ────────────────────────────────────────────
	// Separate group so :id refers to agendaID, not eventID
	agenda := app.Group("/api/events/agenda", middleware.Protected())

	agenda.Put("/:id", middleware.RequireRole("alumni", "student"), controllers.UpdateAgenda)
	agenda.Delete("/:id", middleware.RequireRole("alumni", "student"), controllers.DeleteAgenda)
}
