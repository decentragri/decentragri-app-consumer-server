package middleware

import (
	"os"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupSecurityMiddleware adds comprehensive security middleware to the app
func SetupSecurityMiddleware(app *fiber.App) {
	// Recovery middleware to handle panics
	app.Use(recover.New(recover.Config{
		EnableStackTrace: os.Getenv("NODE_ENV") != "production",
	}))

	// Security headers
	app.Use(helmet.New(helmet.Config{
		XSSProtection:             "1; mode=block",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "DENY",
		ReferrerPolicy:            "no-referrer",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "cross-origin",
		OriginAgentCluster:        "?1",
		XDNSPrefetchControl:       "off",
		XDownloadOptions:          "noopen",
		XPermittedCrossDomain:     "none",
	}))

	// Request logging (only in development)
	if os.Getenv("NODE_ENV") != "production" {
		app.Use(logger.New(logger.Config{
			Format: "[${time}] ${status} - ${method} ${path} - ${ip} - ${latency}\n",
		}))
	}

	// Rate limiting
	app.Use(limiter.New(limiter.Config{
		Max:               100,              // requests
		Expiration:        15 * time.Minute, // per 15 minutes
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("x-forwarded-for", c.IP())
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, please try again later",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
		},
	}))
}

// SetupAPIRateLimit sets up specific rate limiting for API endpoints
func SetupAPIRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:               50,               // requests
		Expiration:        10 * time.Minute, // per 10 minutes
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP + User-Agent for more specific limiting
			return c.Get("x-forwarded-for", c.IP()) + c.Get("User-Agent")
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "API rate limit exceeded",
				"code":  "API_RATE_LIMIT_EXCEEDED",
			})
		},
	})
}
