package routes

import (
	"fmt"

	"decentragri-app-cx-server/middleware"
	portfolioservices "decentragri-app-cx-server/portfolio.services"

	"github.com/gofiber/fiber/v2"
)

func PortfolioRoutes(app *fiber.App, limiter fiber.Handler) {
	api := app.Group("/api")

	// Apply rate limiting to portfolio routes
	api.Use(limiter)

	// Protected portfolio group requiring authentication
	portfolioGroup := api.Group("/portfolio")
	portfolioGroup.Use(middleware.AuthMiddleware())

	portfolioGroup.Get("/summary", func(c *fiber.Ctx) error {
		token := middleware.ExtractToken(c)

		fmt.Println("tae: ", token)
		fmt.Printf("Received portfolio summary request with token\n")

		response, err := portfolioservices.GetPortFolioSummary(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)
	})

	portfolioGroup.Get("/entire", func(c *fiber.Ctx) error {
		token := middleware.ExtractToken(c)

		response, err := portfolioservices.GetEntirePortfolio(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)
	})
}
