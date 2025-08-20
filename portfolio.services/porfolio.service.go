package portfolioservices

import (
	"decentragri-app-cx-server/cache"
	"decentragri-app-cx-server/config"
	"fmt"
	"time"

	tokenServices "decentragri-app-cx-server/token.services"
	walletServices "decentragri-app-cx-server/wallet.services"
)

type PortfolioSummary struct {
	FarmPlotNFTCount int `json:"farmPlotNFTCount"`
}

func GetPortFolioSummary(token string) (PortfolioSummary, error) {
	var username string
	var err error

	// Check if this is a dev bypass token
	if token == "dev_bypass_authorized" {
		fmt.Println("Dev bypass detected in portfolio service")
		username = "0x984785A89BF95cb3d5Df4E45F670081944d8D547" // Use dev wallet address
	} else {
		// Normal token verification
		username, err = tokenServices.NewTokenService().VerifyAccessToken(token)
		if err != nil {
			return PortfolioSummary{}, err
		}
	}

	// Create cache key for portfolio summary
	cacheKey := fmt.Sprintf("portfolio:%s", username)

	// Try to get from cache first
	var cachedSummary PortfolioSummary
	if cache.Exists(cacheKey) {
		err := cache.Get(cacheKey, &cachedSummary)
		if err == nil {
			return cachedSummary, nil
		}
	}

	farmPlotNFTs, err := walletServices.NewWalletService().GetOwnedNFTs(config.CHAIN, config.FarmPlotContractAddress, username)
	if err != nil {
		return PortfolioSummary{}, err
	}

	farmPlotNFTCount := len(farmPlotNFTs.Result)

	summary := PortfolioSummary{
		FarmPlotNFTCount: farmPlotNFTCount,
	}

	// Cache the portfolio summary for 3 minutes
	cache.Set(cacheKey, summary, 3*time.Minute)

	return summary, nil
}
