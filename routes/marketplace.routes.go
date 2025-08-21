package routes

import (
	marketplaceservices "decentragri-app-cx-server/marketplace.services"
	"decentragri-app-cx-server/middleware"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func MarketplaceRoutes(app *fiber.App) {
	group := app.Group("/api/marketplace")

	// Apply auth middleware to all marketplace routes
	group.Use(middleware.AuthMiddleware())

	// GET /api/marketplace/valid-farmplots
	group.Get("/valid-farmplots", func(c *fiber.Ctx) error {
		start := time.Now() // Start timing
		path := c.Path()
		method := c.Method()

		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)

		token := middleware.ExtractToken(c)
		result, err := marketplaceservices.GetValidFarmPlotListings(token)

		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[%s] %s request to %s failed after %s: %v\n",
				time.Now().Format(time.RFC3339), method, path, elapsed, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		fmt.Printf("[%s] Completed %s request to %s successfully in %s\n",
			time.Now().Format(time.RFC3339), method, path, elapsed)
		return c.JSON(result)
	})

	// GET /api/marketplace/featured-property
	group.Get("/featured-property", func(c *fiber.Ctx) error {
		start := time.Now() // Start timing
		path := c.Path()
		method := c.Method()

		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)

		token := middleware.ExtractToken(c)
		result, err := marketplaceservices.FeaturedProperty(token)

		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[%s] %s request to %s failed after %s: %v\n",
				time.Now().Format(time.RFC3339), method, path, elapsed, err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		fmt.Printf("[%s] Completed %s request to %s successfully in %s\n",
			time.Now().Format(time.RFC3339), method, path, elapsed)
		return c.JSON(result)
	})
}
