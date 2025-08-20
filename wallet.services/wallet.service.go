package walletservices

import (
	"decentragri-app-cx-server/cache"
	"decentragri-app-cx-server/config"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetTokenPriceUSD fetches token price using Fiber client
func (is *InsightService) GetTokenPriceUSD(chainID int, tokenAddress string) (float64, error) {
	if tokenAddress == "" {
		tokenAddress = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" // Native token
	}

	// Create cache key for price
	cacheKey := fmt.Sprintf("price:%d:%s", chainID, tokenAddress)

	// Try to get from cache first
	var cachedPrice float64
	if cache.Exists(cacheKey) {
		err := cache.Get(cacheKey, &cachedPrice)
		if err == nil {
			return cachedPrice, nil
		}
	}

	url := fmt.Sprintf("https://%d.insight.thirdweb.com/v1/tokens/price?address=%s", chainID, tokenAddress)

	// Use Fiber's client instead of net/http
	req := fiber.Get(url)
	req.Set("x-secret-key", is.secretKey)

	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return 0, fmt.Errorf("failed to make request: %v", errs[0])
	}

	if status < 200 || status >= 300 {
		return 0, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	var priceResp PriceResponse
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(priceResp.Data) == 0 {
		return 0, errors.New("price data not found")
	}

	price := priceResp.Data[0].PriceUSD

	// Cache the price for 2 minutes (prices change frequently)
	cache.Set(cacheKey, price, 2*time.Minute)

	return price, nil
}

// SafeGetPrice safely gets token price, returns 0 on error
func (is *InsightService) SafeGetPrice(chainID int, tokenAddress string) float64 {
	price, err := is.GetTokenPriceUSD(chainID, tokenAddress)
	if err != nil {
		fmt.Printf("Failed to fetch price for %s on chain %d: %v\n", tokenAddress, chainID, err)
		return 0
	}
	return price
}

// GetBalance fetches native token balance using Fiber client
func (ws *WalletService) GetBalance(chainID, walletAddress string) (BalanceResponse, error) {
	url := fmt.Sprintf("%s/backend-wallet/%s/get-balance?walletAddress=%s", config.EngineCloudBaseURL, chainID, walletAddress)

	// Use Fiber's client instead of net/http
	req := fiber.Get(url)
	req.Set("x-secret-key", ws.secretKey)

	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return BalanceResponse{}, fmt.Errorf("failed to make request: %v", errs[0])
	}

	if status < 200 || status >= 300 {
		return BalanceResponse{}, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	var balanceResp BalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return BalanceResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return balanceResp, nil
}

// GetERC20Balance fetches ERC20 token balance using Fiber client
func (ws *WalletService) GetERC20Balance(chainID, walletAddress, tokenAddress string) (BalanceResponse, error) {
	url := fmt.Sprintf("%s/contract/%s/%s/erc20/balance-of?walletAddress=%s", config.EngineCloudBaseURL, chainID, tokenAddress, walletAddress)

	// Use Fiber's client instead of net/http
	req := fiber.Get(url)
	req.Set("x-secret-key", ws.secretKey)

	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return BalanceResponse{}, fmt.Errorf("failed to make request: %v", errs[0])
	}

	if status < 200 || status >= 300 {
		return BalanceResponse{}, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	var balanceResp BalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return BalanceResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return balanceResp, nil
}

// GetWalletBalance fetches wallet balance and price data using Fiber client
func (ws *WalletService) GetWalletBalance(walletAddress string) (WalletData, error) {
	insightService := NewInsightService()

	// Convert CHAIN to int for price calls
	chainInt, err := strconv.Atoi(CHAIN)
	if err != nil {
		return WalletData{}, fmt.Errorf("invalid chain ID: %w", err)
	}

	// Use goroutines and WaitGroup for concurrent API calls
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Results storage
	var (
		ethToken    BalanceResponse
		swellToken  BalanceResponse
		dagriToken  BalanceResponse
		rsWETH      BalanceResponse
		dagriPrice  float64
		ethPrice    float64
		swellPrice  float64
		fetchErrors []error
	)

	// Helper function to handle errors safely
	addError := func(err error) {
		if err != nil {
			mu.Lock()
			fetchErrors = append(fetchErrors, err)
			mu.Unlock()
		}
	}

	// Fetch balances concurrently
	wg.Add(7)

	// ETH balance
	go func() {
		defer wg.Done()
		balance, err := ws.GetBalance("1", walletAddress)
		mu.Lock()
		ethToken = balance
		mu.Unlock()
		addError(err)
	}()

	// Swell balance
	go func() {
		defer wg.Done()
		balance, err := ws.GetBalance(CHAIN, walletAddress)
		mu.Lock()
		swellToken = balance
		mu.Unlock()
		addError(err)
	}()

	// DAGRI token balance
	go func() {
		defer wg.Done()
		balance, err := ws.GetERC20Balance(CHAIN, walletAddress, DECENTRAGRI_TOKEN)
		mu.Lock()
		dagriToken = balance
		mu.Unlock()
		addError(err)
	}()

	// rsWETH balance
	go func() {
		defer wg.Done()
		balance, err := ws.GetERC20Balance("1", walletAddress, RSWETH_ADDRESS)
		mu.Lock()
		rsWETH = balance
		mu.Unlock()
		addError(err)
	}()

	// DAGRI price
	go func() {
		defer wg.Done()
		price := insightService.SafeGetPrice(chainInt, DECENTRAGRI_TOKEN)
		mu.Lock()
		dagriPrice = price
		mu.Unlock()
	}()

	// ETH price
	go func() {
		defer wg.Done()
		price := insightService.SafeGetPrice(1, "")
		mu.Lock()
		ethPrice = price
		mu.Unlock()
	}()

	// SWELL price
	go func() {
		defer wg.Done()
		price := insightService.SafeGetPrice(1, "0x0a6E7Ba5042B38349e437ec6Db6214AEC7B35676")
		mu.Lock()
		swellPrice = price
		mu.Unlock()
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Check if there were critical errors (balance fetching errors)
	if len(fetchErrors) > 0 {
		return WalletData{}, fmt.Errorf("failed to fetch wallet data: %v", fetchErrors)
	}

	return WalletData{
		SmartWalletAddress: walletAddress,

		// Balances
		EthBalance:    ethToken.Result.DisplayValue,
		SwellBalance:  swellToken.Result.DisplayValue,
		RsWETHBalance: rsWETH.Result.DisplayValue,
		DagriBalance:  dagriToken.Result.DisplayValue,
		NativeBalance: swellToken.Result.DisplayValue,

		// Prices
		DagriPriceUSD: dagriPrice,
		EthPriceUSD:   ethPrice,
		SwellPriceUSD: swellPrice,
	}, nil
}

// GetOwnedNFTs fetches owned NFTs from a specific contract using Fiber client
func (ws *WalletService) GetOwnedNFTs(chainID, contractAddress, walletAddress string) (NFTResponse, error) {
	url := fmt.Sprintf("%s/contract/%s/%s/erc1155/get-owned?walletAddress=%s", config.EngineCloudBaseURL, chainID, contractAddress, walletAddress)
	println("Fetching NFTs from URL:", url)
	// Use Fiber's client instead of net/http
	req := fiber.Get(url)
	req.Set("x-secret-key", ws.secretKey)
	req.Set("Authorization", "Bearer "+ws.secretKey)

	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return NFTResponse{}, fmt.Errorf("failed to make request: %v", errs[0])
	}

	if status < 200 || status >= 300 {
		return NFTResponse{}, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	var nftResp NFTResponse
	if err := json.Unmarshal(body, &nftResp); err != nil {
		return NFTResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return nftResp, nil
}
