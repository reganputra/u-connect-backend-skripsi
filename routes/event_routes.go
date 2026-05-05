package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/controllers"
	"github.com/reganputra/skripsi-backend/middleware"
	"gorm.io/gorm"
)

func RegisterEventRoutes(app *fiber.App, ctrl *controllers.EventController, db *gorm.DB) {
	// ── Events (/api/events) ──────────────────────────────────────────────────
	events := app.Group("/api/events", middleware.Protected())

	events.Get("/", ctrl.GetEvents)
	events.Get("/mine/owned", middleware.RequireRole("alumni", "student"), ctrl.GetMyOwnedEvents)
	events.Get("/mine/registered", middleware.RequireRole("alumni", "student"), ctrl.GetMyRegisteredEvents)
	events.Post("/", middleware.RequireRole("alumni", "student"), ctrl.CreateEvent)
	events.Get("/:id", middleware.TrackView("event", db), ctrl.GetEventByID)
	events.Put("/:id", middleware.RequireRole("alumni", "student"), ctrl.UpdateEvent)
	events.Delete("/:id", middleware.RequireRole("alumni", "student"), ctrl.DeleteEvent)

	// Registration
	events.Post("/:id/register", middleware.RequireRole("alumni", "student"), ctrl.RegisterForEvent)
	events.Delete("/:id/register", middleware.RequireRole("alumni", "student"), ctrl.CancelRegistration)
	events.Get("/:id/participants", ctrl.GetParticipants)

	// Agenda
	events.Post("/:id/agenda", middleware.RequireRole("alumni", "student"), ctrl.AddAgenda)

	// ── Agenda (/api/agenda) ──────────────────────────────────────────────────
	// Separate group to avoid conflicts with /:id param
	agenda := app.Group("/api/agenda", middleware.Protected())

	agenda.Put("/:id", middleware.RequireRole("alumni", "student"), ctrl.UpdateAgenda)
	agenda.Delete("/:id", middleware.RequireRole("alumni", "student"), ctrl.DeleteAgenda)
}
