package main

import (
	"decentragri-app-cx-server/cache"
	memgraph "decentragri-app-cx-server/db"
	"decentragri-app-cx-server/routes"
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	log.Printf("Server configured to use %d CPU cores", runtime.NumCPU())

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	// Initialize Memgraph database connection
	memgraph.InitMemGraph()

	// Initialize Redis
	cache.InitRedis()

	app := fiber.New(fiber.Config{
		AppName:      "Decentragri App CX Server",
		ServerHeader: "Decentragri App CX Server",
		BodyLimit:    50 * 1024 * 1024, //50 MB
		IdleTimeout:  60,
	})

	routes.AuthRoutes(app)
	routes.PortfolioRoutes(app)
	routes.MarketplaceRoutes(app)
	routes.WalletRoutes(app)

	if err := app.Listen(":9085"); err != nil {
		panic(err)
	}
}
