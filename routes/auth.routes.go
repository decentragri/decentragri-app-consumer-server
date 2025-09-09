package routes

import (
	authservices "decentragri-app-cx-server/auth.services"
	memgraph "decentragri-app-cx-server/db"
	tokenServices "decentragri-app-cx-server/token.services"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App, limiter fiber.Handler) {
	authGroup := app.Group("/api")

	// Apply rate limiting to auth routes
	authGroup.Use(limiter)

	//** WALLET AUTHENTICATION ROUTES **//
	authGroup.Post("/auth/nonce", func(c *fiber.Ctx) error {
		var req authservices.GetNonceRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		fmt.Printf("Received nonce request: %+v\n", req)

		response, err := authservices.GetNonce(req.WalletAddress)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)
	})

	authGroup.Post("/auth/authenticate/wallet", func(c *fiber.Ctx) error {
		var req authservices.AuthenticateWalletRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		fmt.Printf("Received authentication data: %+v\n", req)

		response, err := authservices.AuthenticateWallet(req)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)
	})

	//** DEV BYPASS ROUTE - REMOVE IN PRODUCTION **//
	authGroup.Post("/auth/dev-bypass", func(c *fiber.Ctx) error {
		// Check if dev bypass is enabled
		if !authservices.CheckDevBypass(c) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Dev bypass not enabled"})
		}

		fmt.Println("Dev bypass authentication used")

		// Use a dev user wallet address
		devWalletAddress := "0x984785A89BF95cb3d5Df4E45F670081944d8D547"

		// Check if dev user exists, create if not
		query := `MATCH (u:User {username: $username}) RETURN u.username AS username`
		params := map[string]any{"username": devWalletAddress}
		records, err := memgraph.ExecuteRead(query, params)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}

		// Create dev user if it doesn't exist
		if len(records) == 0 {
			createQuery := `CREATE (u:User {
				username: $username,
				createdAt: timestamp(),
				walletAddress: $walletAddress,
				deviceId: $deviceId,
				authProvider: 'dev_bypass'
			}) RETURN u.username AS username`
			createParams := map[string]any{
				"username":      devWalletAddress,
				"walletAddress": devWalletAddress,
				"deviceId":      "dev_device_001",
			}
			_, err = memgraph.ExecuteWrite(createQuery, createParams)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create dev user: " + err.Error()})
			}
			fmt.Println("Dev user created in database")
		}

		// Generate tokens for the dev user
		tokenService := tokenServices.NewTokenService()
		tokens, err := tokenService.GenerateTokens(devWalletAddress)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate dev tokens"})
		}

		response := authservices.AuthenticateWalletResponse{
			WalletAddress: devWalletAddress,
			Tokens:        *tokens,
			IsNewUser:     len(records) == 0,
			Message:       "Dev bypass authentication successful",
			LoginType:     "dev_bypass",
		}

		return c.JSON(response)
	})

	//** GOOGLE AUTHENTICATION ROUTES **//
	authGroup.Post("/auth/authenticate/google", func(c *fiber.Ctx) error {
		var req authservices.AuthenticateGoogleRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		fmt.Printf("Received Google authentication request: Device ID: %s\n", req.DeviceId)

		response, err := authservices.AuthenticateGoogle(req)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(response)
	})

	authGroup.Post("/renew/access/decentra", func(c *fiber.Ctx) error {
		var req authservices.RefreshTokenRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}

		fmt.Printf("Received refresh token request: %+v\n", req)

		tokens, err := authservices.RefreshSession(req.RefreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(tokens)
	})

}
