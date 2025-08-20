package middleware

import (
	authservices "decentragri-app-cx-server/auth.services"
	tokenServices "decentragri-app-cx-server/token.services"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT tokens or allows dev bypass
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Printf("Auth middleware - Path: %s\n", c.Path())

		// Check for dev bypass first
		if authservices.CheckDevBypass(c) {
			fmt.Println("Dev bypass activated - allowing access")
			// Just set minimal required context and allow access
			c.Locals("isDev", true)
			c.Locals("username", "dev_user")
			return c.Next()
		}

		fmt.Println("Dev bypass not activated, checking JWT token...")

		// Extract token from Authorization header
		token := c.Get("Authorization")
		if token == "" {
			fmt.Println("No Authorization header found")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		fmt.Printf("Validating JWT token: %s...\n", token[:20])

		// Validate the token
		tokenService := tokenServices.NewTokenService()
		username, err := tokenService.VerifyAccessToken(token)
		if err != nil {
			fmt.Printf("JWT validation failed: %v\n", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		fmt.Printf("JWT validation successful for user: %s\n", username)

		// Store user info in context for use in handlers
		c.Locals("username", username)
		c.Locals("isDev", false)

		return c.Next()
	}
}

// ExtractToken helper function for routes that need the raw token
func ExtractToken(c *fiber.Ctx) string {
	// Check if this is a dev bypass request
	if isDev, ok := c.Locals("isDev").(bool); ok && isDev {
		fmt.Println("Dev bypass - returning dummy token for services")
		return "dev_bypass_authorized" // Simple placeholder that indicates dev bypass
	}

	// Extract real token from Authorization header for normal authentication
	token := c.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	return token
}
