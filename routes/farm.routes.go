package routes

import (
	"log"

	farmservices "decentragri-app-cx-server/farm.services"
	"decentragri-app-cx-server/utils"

	"github.com/gofiber/fiber/v2"
)

func FarmRoutes(app *fiber.App, limiter fiber.Handler) {
	api := app.Group("/api")

	// Apply rate limiting to farm routes
	api.Use(limiter)

	// Define farm group for farm-specific routes
	farmGroup := api.Group("/farm")

	// GET /api/farm/list - Get user's farms with formatted dates and image bytes
	farmGroup.Get("/list", func(c *fiber.Ctx) error {
		// token := middleware.ExtractToken(c)

		log.Println("Processing farm list request")

		response, err := farmservices.GetFarmList()
		if err != nil {
			log.Printf("Error fetching farm list: %v", err)
			return utils.HandleInternalError(c, err, "fetching farm list")
		}

		return c.JSON(response)
	})

	// GET /api/farm/scans/:farmName - Get recent farm scans with pagination
	farmGroup.Get("/scans/:farmName", func(c *fiber.Ctx) error {
		farmName := utils.SanitizeInput(c.Params("farmName"))

		// Validate farm name input
		if !utils.ValidateFarmName(farmName) {
			log.Printf("Invalid farm name provided: %s", farmName)
			return utils.HandleValidationError(c, "farmName")
		}

		// Get pagination parameters with validation
		page, limit, err := utils.ValidatePagination(c.Query("page"), c.Query("limit"))
		if err != nil {
			return utils.HandleValidationError(c, err.Error())
		}

		log.Printf("Processing farm scans request for farm: %s, page: %d, limit: %d", farmName, page, limit)

		response, err := farmservices.GetFarmScans(farmName, page, limit)
		if err != nil {
			log.Printf("Error fetching farm scans: %v", err)
			return utils.HandleInternalError(c, err, "fetching farm scans")
		}

		return c.JSON(response)
	})
}
