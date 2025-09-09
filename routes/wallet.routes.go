// Package routes provides HTTP route handlers for the Decentragri wallet API endpoints.
// This package defines RESTful API routes for wallet management operations including
// wallet creation, balance queries, and NFT ownership verification.
//
// Route Configuration:
//   - Base path: /api/wallet
//   - Authentication: Required for all endpoints (JWT middleware)
//   - Request logging: Comprehensive timing and error logging
//   - Error handling: Standardized JSON error responses
//
// Supported Operations:
//   - POST /api/wallet/create: Create new smart wallets
//   - GET /api/wallet/balances: Retrieve comprehensive token balances
//   - GET /api/wallet/nfts/:contract: Query NFT ownership from specific contracts
//
// Security Features:
//   - JWT authentication middleware on all routes
//   - Automatic token extraction and validation
//   - Request timing and audit logging
//   - Error sanitization to prevent information leakage
package routes

import (
	"decentragri-app-cx-server/middleware"
	walletServices "decentragri-app-cx-server/wallet.services"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// WalletRoutes configures and registers all wallet-related HTTP endpoints.
// This function sets up the wallet API route group with authentication middleware
// and comprehensive request logging for monitoring and debugging purposes.
//
// Route Group Configuration:
//   - Base path: /api/wallet
//   - Middleware: JWT authentication required for all routes
//   - Logging: Request timing, method, path, and outcome tracking
//   - Error handling: Consistent JSON error response format
//
// Registered Endpoints:
//   - POST /create: Smart wallet creation with ThirdWeb integration
//   - GET /balances: Multi-token balance queries with USD pricing
//   - GET /nfts/:contract: NFT ownership queries for specific contracts
//
// Performance Monitoring:
//   - Request start time tracking
//   - Execution duration measurement
//   - Success/failure outcome logging
//   - Detailed error reporting for debugging
//
// Parameters:
//   - app: The Fiber application instance to register routes with
//
// Security Implementation:
//   - JWT token validation on all routes
//   - Automatic wallet address extraction from tokens
//   - Request authentication status logging
//   - Protected resource access control
func WalletRoutes(app *fiber.App, limiter fiber.Handler) {
	// Initialize wallet service for handling wallet operations
	walletService := walletServices.NewWalletService()

	// Create wallet API route group with rate limiting and authentication middleware
	wallet := app.Group("/api/wallet")
	wallet.Use(limiter)
	wallet.Use(middleware.AuthMiddleware())

	// POST /api/wallet/create - Create new smart wallet
	// This endpoint creates a new ThirdWeb smart wallet for the authenticated user
	// Authentication: JWT token required
	// Response: Wallet creation confirmation with address
	wallet.Post("/create", func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)

		// Extract JWT token for user identification
		token := middleware.ExtractToken(c)

		// Create wallet using the ThirdWeb service
		walletResponse, err := walletService.CreateWallet(token)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[%s] %s request to %s failed after %s: %v\n", time.Now().Format(time.RFC3339), method, path, elapsed, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		fmt.Printf("[%s] Completed %s request to %s successfully in %s\n", time.Now().Format(time.RFC3339), method, path, elapsed)
		return c.Status(fiber.StatusCreated).JSON(walletResponse)
	})

	// GET /api/wallet/balances - Retrieve comprehensive token balances
	// This endpoint fetches native and ERC20 token balances with USD pricing
	// Authentication: JWT token required
	// Response: Complete balance information with USD values
	wallet.Get("/balances", func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)

		// Extract JWT token for user identification
		token := middleware.ExtractToken(c)

		// Fetch comprehensive user balance information
		balances, err := walletService.GetUserBalances(token)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[%s] %s request to %s failed after %s: %v\n", time.Now().Format(time.RFC3339), method, path, elapsed, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		fmt.Printf("[%s] Completed %s request to %s successfully in %s\n", time.Now().Format(time.RFC3339), method, path, elapsed)
		return c.JSON(balances)
	})

	// GET /api/wallet/nfts/:contract - Query NFT ownership from specific contracts
	// This endpoint retrieves all NFTs owned by the user from a specified contract
	// Authentication: JWT token required
	// Parameters: contract (path) - The contract address to query
	// Response: Array of owned NFTs with metadata and quantities
	wallet.Get("/nfts/:contract", func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)

		// Extract contract address from URL parameters
		contract := c.Params("contract")
		if contract == "" {
			fmt.Printf("[%s] %s request to %s failed: contract parameter is required\n", time.Now().Format(time.RFC3339), method, path)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "contract parameter is required"})
		}

		// Extract JWT token for user identification
		token := middleware.ExtractToken(c)

		// Fetch NFT ownership data for the specified contract
		nfts, err := walletService.GetOwnedNFTs(contract, token)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[%s] %s request to %s failed after %s: %v\n", time.Now().Format(time.RFC3339), method, path, elapsed, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		fmt.Printf("[%s] Completed %s request to %s successfully in %s\n", time.Now().Format(time.RFC3339), method, path, elapsed)
		return c.JSON(nfts)
	})
}
