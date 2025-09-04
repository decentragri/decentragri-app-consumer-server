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
	"runtime"

	"github.com/gofiber/fiber/v2"
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
	// Configure the Go runtime to use all available CPU cores for maximum performance.
	// This enables concurrent processing of HTTP requests and background tasks.
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Printf("Server configured to use %d CPU cores for optimal performance", runtime.NumCPU())

	// Load environment variables from .env file for configuration.
	// This includes database credentials, API keys, and other sensitive configuration.
	// The server will continue with system environment variables if .env file is not found.
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file, using system environment variables:", err)
	} else {
		log.Println("Environment variables loaded successfully")
	}

	// Initialize Memgraph database connection for graph-based data storage.
	// Memgraph is used for storing user relationships, transaction history, and complex queries.
	log.Println("Initializing Memgraph database connection...")
	memgraph.InitMemGraph()
	log.Println("Memgraph database connected successfully")

	// Initialize Redis cache for session management and performance optimization.
	// Redis is used for JWT token storage, API response caching, and temporary data.
	log.Println("Initializing Redis cache...")
	cache.InitRedis()
	log.Println("Redis cache connected successfully")

	// Create a new Fiber application instance with custom configuration for optimal performance.
	app := fiber.New(fiber.Config{
		AppName:      "Decentragri App CX Server", // Application identifier
		ServerHeader: "Decentragri App CX Server", // HTTP server header
		BodyLimit:    50 * 1024 * 1024,            // 50 MB request body limit for file uploads
		IdleTimeout:  60,                          // 60 seconds idle timeout for connections
		Prefork:      false,                       // Disabled for development (enable for production)
		ReadTimeout:  30,                          // 30 seconds read timeout
		WriteTimeout: 30,                          // 30 seconds write timeout
	})

	log.Println("Registering API routes...")

	// Register all API route groups with their respective middleware and handlers.
	// Each route group handles a specific domain of functionality:

	// Authentication routes: login, token refresh, user management
	routes.AuthRoutes(app)
	log.Println("  Authentication routes registered")

	// Portfolio routes: NFT management, farm plot listings, user portfolios
	routes.PortfolioRoutes(app)
	log.Println("  Portfolio routes registered")

	// Marketplace routes: buy/sell operations, featured properties, listings
	routes.MarketplaceRoutes(app)
	log.Println("  Marketplace routes registered")

	// Wallet routes: balance queries, NFT ownership, wallet creation
	routes.WalletRoutes(app)
	log.Println("  Wallet routes registered")

	// Farm routes: farm listings, user farms, farm management
	routes.FarmRoutes(app)
	log.Println("  Farm routes registered")

	log.Println("All routes registered successfully")

	// Start the HTTP server on port 9085.
	// The server will listen for incoming requests and handle them concurrently.
	// If the server fails to start (e.g., port already in use), the application will panic.
	log.Println("Starting HTTP server on port 9085...")
	log.Println("Server endpoints available at: http://localhost:9085")
	log.Println("API documentation: Check README.md for available endpoints")

	if err := app.Listen(":9085"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		panic(err)
	}
}
