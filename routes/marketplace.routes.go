package routes

import (
	marketplaceservices "decentragri-app-cx-server/marketplace.services"
	"decentragri-app-cx-server/middleware"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func MarketplaceRoutes(app *fiber.App, limiter fiber.Handler) {
	api := app.Group("/api")

	// Apply rate limiting to marketplace routes
	api.Use(limiter)

	// Protected marketplace group requiring authentication
	group := api.Group("/marketplace")
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

	// POST /api/marketplace/buy-from-listing
	group.Post("/buy-from-listing", func(c *fiber.Ctx) error {
		start := time.Now() // Start timing
		path := c.Path()
		method := c.Method()

		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)

		var req marketplaceservices.BuyFromListingRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}

		token := middleware.ExtractToken(c)
		result, err := marketplaceservices.BuyFromListing(token, &req)
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
