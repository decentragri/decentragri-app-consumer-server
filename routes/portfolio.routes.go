package routes

import (
	"fmt"

	"decentragri-app-cx-server/middleware"
	portfolioservices "decentragri-app-cx-server/portfolio.services"

	"github.com/gofiber/fiber/v2"
)

func PortfolioRoutes(app *fiber.App) {
	portfolioGroup := app.Group("/api/portfolio")

	// Apply auth middleware to all portfolio routes
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
}
