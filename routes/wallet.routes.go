package routes

import (
	"decentragri-app-cx-server/middleware"
	walletServices "decentragri-app-cx-server/wallet.services"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func WalletRoutes(app *fiber.App) {
	walletService := walletServices.NewWalletService()
	wallet := app.Group("/api/wallet")
	wallet.Use(middleware.AuthMiddleware())

	// POST /api/wallet/create
	wallet.Post("/create", func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)


		wallet, err := walletService.CreateWallet()
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[%s] %s request to %s failed after %s: %v\n", time.Now().Format(time.RFC3339), method, path, elapsed, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		fmt.Printf("[%s] Completed %s request to %s successfully in %s\n", time.Now().Format(time.RFC3339), method, path, elapsed)
		return c.Status(fiber.StatusCreated).JSON(wallet)
	})

	// GET /api/wallet/balances
	wallet.Get("/balances", func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)

		token := middleware.ExtractToken(c)
		balances, err := walletService.GetUserBalances(token)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[%s] %s request to %s failed after %s: %v\n", time.Now().Format(time.RFC3339), method, path, elapsed, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		fmt.Printf("[%s] Completed %s request to %s successfully in %s\n", time.Now().Format(time.RFC3339), method, path, elapsed)
		return c.JSON(balances)
	})

	// GET /api/wallet/nfts/:contract
	wallet.Get("/nfts/:contract", func(c *fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		method := c.Method()
		fmt.Printf("[%s] Starting %s request to %s\n", start.Format(time.RFC3339), method, path)

		contract := c.Params("contract")
		token := middleware.ExtractToken(c)
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
