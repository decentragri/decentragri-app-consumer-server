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
	"decentragri-app-cx-server/middleware"
	"decentragri-app-cx-server/routes"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/joho/godotenv"
)

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
		// Security: Disable server header in production
		DisableStartupMessage: os.Getenv("NODE_ENV") == "production",
		// Enable proxy support for proper IP detection behind Nginx
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		ProxyHeader:             "X-Forwarded-For",
		// Error handling
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			log.Printf("Fiber error (%d): %v", code, err)

			return c.Status(code).JSON(fiber.Map{
				"error": "An error occurred processing your request",
				"code":  code,
			})
		},
	})

	// Setup security middleware
	middleware.SetupSecurityMiddleware(app)

	// Configure rate limiting to prevent abuse with proxy-aware IP detection
	rateLimiter := limiter.New(limiter.Config{
		Max:        30,              // 30 requests per window
		Expiration: 1 * time.Minute, // 1 minute window
		KeyGenerator: func(c *fiber.Ctx) string {
			// Get real client IP, handling proxy headers
			clientIP := c.IP()

			// Check for forwarded IP headers (for Nginx proxy)
			if forwardedFor := c.Get("X-Forwarded-For"); forwardedFor != "" {
				// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
				// Take the first one (original client)
				if parts := strings.Split(forwardedFor, ","); len(parts) > 0 {
					clientIP = strings.TrimSpace(parts[0])
				}
			} else if realIP := c.Get("X-Real-IP"); realIP != "" {
				// Alternative header used by some proxies
				clientIP = realIP
			}

			log.Printf("Rate limiting key for IP: %s", clientIP)
			return clientIP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
			})
		},
	})

	// Add CORS middleware with security-focused configuration

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Environment-driven origins for security
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Dev-Bypass-Token",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowCredentials: false, // Enable credentials for authenticated requests
	}))

	routes.AuthRoutes(app, rateLimiter)
	routes.PortfolioRoutes(app, rateLimiter)
	routes.MarketplaceRoutes(app, rateLimiter)
	routes.WalletRoutes(app, rateLimiter)
	routes.FarmRoutes(app, rateLimiter)

	// Configure server with environment-driven settings
	port := os.Getenv("PORT")
	if port == "" {
		port = "9085" // Default port
	}

	log.Printf("Starting HTTP server on port %s...", port)
	log.Printf("Server endpoints available at: http://localhost:%s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
