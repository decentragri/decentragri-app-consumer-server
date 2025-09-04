package routes

import (
	"fmt"

	farmservices "decentragri-app-cx-server/farm.services"
	// "decentragri-app-cx-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func FarmRoutes(app *fiber.App) {
	farmGroup := app.Group("/api")

	// Apply auth middleware to all farm routes
	// farmGroup.Use(middleware.AuthMiddleware())

	// GET /api/farm/list - Get user's farms with formatted dates and image bytes
	farmGroup.Get("/farm/list", func(c *fiber.Ctx) error {
		// token := middleware.ExtractToken(c)

		fmt.Printf("Received farm list request with token\n")

		response, err := farmservices.GetFarmList()
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)
	})
}
