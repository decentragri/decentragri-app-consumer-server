package routes

import (
	walletServices "decentragri-app-cx-server/wallet.services"

	"github.com/gofiber/fiber/v2"
)

func WalletRoutes(app *fiber.App) {
	walletService := walletServices.NewWalletService()

	wallet := app.Group("/api/wallet")

	// Create new wallet
	wallet.Post("/create", func(c *fiber.Ctx) error {
		wallet, err := walletService.CreateWallet()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusCreated).JSON(wallet)
	})

	// Get user balances
	wallet.Get("/balances/:chainId/:address", func(c *fiber.Ctx) error {
		chainID := c.Params("chainId")
		address := c.Params("address")

		balances, err := walletService.GetUserBalances(chainID, address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(balances)
	})

	// Get owned NFTs
	wallet.Get("/nfts/:chainId/:contract/:address", func(c *fiber.Ctx) error {
		chainID := c.Params("chainId")
		contract := c.Params("contract")
		address := c.Params("address")

		nfts, err := walletService.GetOwnedNFTs(chainID, contract, address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(nfts)
	})
}
