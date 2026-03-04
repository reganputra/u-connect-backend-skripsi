package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/reganputra/skripsi-backend/config"
	"github.com/reganputra/skripsi-backend/utils"
)

func main() {
	// Load environment variables from .env file
	utils.LoadEnvFile(".env")

	// Connect to the database
	config.ConnectDB()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Alumni Community Platform API v1.0",
	})

	// Middleware
	app.Use(logger.New())

	// Health check route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// Get port from env, default to 8080
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
