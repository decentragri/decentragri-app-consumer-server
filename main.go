// Package main is the entry point for the Decentragri App CX Server.
// This server provides REST API endpoints for authentication, wallet management,
// marketplace functionality, and portfolio services for blockchain-based agricultural NFTs.
//
// The server utilizes:
//   - Fiber web framework for high-performance HTTP handling
//   - Memgraph for graph-based data storage
//   - Redis for caching and session management
//   - JWT for secure authentication
//   - Multi-core processing for optimal performance
//
// Author: Decentragri Core Team
// Version: 1.0.0
package main

import (
	"decentragri-app-cx-server/cache"
	memgraph "decentragri-app-cx-server/db"
	"decentragri-app-cx-server/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

// main is the application entry point that initializes all services and starts the HTTP server.
// It performs the following operations in order:
//  1. Configures multi-core processing
//  2. Loads environment variables
//  3. Initializes database connections (Memgraph and Redis)
//  4. Sets up Fiber web server with custom configuration
//  5. Registers all API routes
//  6. Starts the HTTP server on port 9085
//
// The server will panic if it fails to start, ensuring no partial initialization states.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file, using system environment variables:", err)
	} else {
		log.Println("Environment variables loaded successfully")
	}

	memgraph.InitMemGraph()
	cache.InitRedis()

	app := fiber.New(fiber.Config{
		AppName:      "Decentragri App CX Server", // Application identifier
		ServerHeader: "Decentragri App CX Server", // HTTP server header
		BodyLimit:    50 * 1024 * 1024,            // 50 MB request body limit for file uploads
		Prefork:      false,
	})

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Allow all origins for development
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Dev-Bypass-Token",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowCredentials: false, // Set to false when using wildcard origins
	}))

	routes.AuthRoutes(app)
	routes.PortfolioRoutes(app)
	routes.MarketplaceRoutes(app)
	routes.WalletRoutes(app)
	routes.FarmRoutes(app)

	log.Println("Starting HTTP server on port 9085...")
	log.Println("Server endpoints available at: http://localhost:9085")

	if err := app.Listen(":9085"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		panic(err)
	}
}
