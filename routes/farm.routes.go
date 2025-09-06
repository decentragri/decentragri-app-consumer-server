package routes

import (
	"fmt"
	"strconv"

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

	// GET /api/farm/scans/:farmName - Get recent farm scans with pagination
	farmGroup.Get("/farm/scans/:farmName", func(c *fiber.Ctx) error {
		farmName := c.Params("farmName")

		// Get pagination parameters from query string
		page := 1
		limit := 10

		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 { // Max 100 items per page
				limit = l
			}
		}

		fmt.Printf("Received farm scans request for farm: %s, page: %d, limit: %d\n", farmName, page, limit)

		response, err := farmservices.GetFarmScans(farmName, page, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)
	})
}
