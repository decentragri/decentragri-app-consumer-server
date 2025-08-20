package routes

import (
	marketplaceservices "decentragri-app-cx-server/marketplace.services"
	"decentragri-app-cx-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func MarketplaceRoutes(app *fiber.App) {
	group := app.Group("/api/marketplace")

	// Apply auth middleware to all marketplace routes
	group.Use(middleware.AuthMiddleware())

	// GET /api/marketplace/valid-farmplots
	group.Get("/valid-farmplots", func(c *fiber.Ctx) error {
		token := middleware.ExtractToken(c)
		result, err := marketplaceservices.GetValidFarmPlotListings(token)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	})

	// GET /api/marketplace/featured-property
	group.Get("/featured-property", func(c *fiber.Ctx) error {
		token := middleware.ExtractToken(c)
		result, err := marketplaceservices.FeaturedProperty(token)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	})
}
