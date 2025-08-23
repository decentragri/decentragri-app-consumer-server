package walletservices

import (
	"decentragri-app-cx-server/config"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type WalletService struct {
	secretKey string
}

func NewWalletService() *WalletService {
	return &WalletService{
		secretKey: os.Getenv("SECRET_KEY"),
	}
}

// CreateWallet creates a new wallet using thirdweb's backend wallet API
func (ws *WalletService) CreateWallet() (*CreateWalletResponse, error) {
	url := fmt.Sprintf("%s/backend-wallet/create", config.EngineCloudBaseURL)

	reqBody := CreateWalletRequest{
		Type: "smart:local",
	}

	req := fiber.Post(url)
	req.Set("Content-Type", "application/json")
	req.Set("Authorization", fmt.Sprintf("Bearer %s", ws.secretKey))
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}
	req.Body(bodyBytes)

	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return nil, fmt.Errorf("error making request: %v", errs[0])
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	var response CreateWalletResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &response, nil
}

// GetBalance fetches native token balance using Fiber client
func GetBalance(chainID, walletAddress string) (BalanceResponse, error) {
	url := fmt.Sprintf("%s/backend-wallet/%s/get-balance?walletAddress=%s",
		config.EngineCloudBaseURL,
		chainID,
		walletAddress,
	)

	// Create the request using Fiber's client
	req := fiber.Get(url)
	req.Set("Authorization", "Bearer "+os.Getenv("SECRET_KEY"))

	// Send the request
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

// GetTokenPriceUSD fetches token price using Fiber client
func GetTokenPriceUSD(chainID int, tokenAddress string) (float64, error) {
	if tokenAddress == "" {
		tokenAddress = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" // Native token
	}

	url := fmt.Sprintf("https://%d.insight.thirdweb.com/v1/tokens/price?address=%s", chainID, tokenAddress)

	// Create the request using Fiber's client
	req := fiber.Get(url)
	req.Set("x-secret-key", os.Getenv("SECRET_KEY"))

	// Send the request
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
		return 0, fmt.Errorf("no price data available")
	}

	return priceResp.Data[0].PriceUSD, nil
}

// GetUserBalances fetches comprehensive balance information for a user
// including native token balance and price
func (ws *WalletService) GetUserBalances(chainID string, walletAddress string) (*UserBalances, error) {
	// Convert chain ID for price fetching
	chainInt, err := strconv.Atoi(chainID)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	// Get native token balance
	nativeBalance, err := GetBalance(chainID, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch native balance: %w", err)
	}

	// Get native token (ETH) price
	nativePrice, err := GetTokenPriceUSD(chainInt, "")
	if err != nil {
		nativePrice = 0 // Set price to 0 if unable to fetch
	}

	// Calculate USD value
	rawBalance := nativeBalance.Result.Value
	balanceFloat, _ := strconv.ParseFloat(rawBalance, 64)
	valueUSD := balanceFloat * nativePrice

	// Create response
	result := &UserBalances{
		WalletAddress: walletAddress,
		Native: TokenBalance{
			Balance:    nativeBalance.Result.DisplayValue,
			RawBalance: nativeBalance.Result.Value,
			PriceUSD:   nativePrice,
			ValueUSD:   valueUSD,
		},
		LastUpdated: time.Now().Unix(),
	}

	return result, nil
}

// GetOwnedNFTs fetches owned NFTs from a specific contract using Fiber client
func (ws *WalletService) GetOwnedNFTs(chainID, contractAddress, walletAddress string) (NFTResponse, error) {
	url := fmt.Sprintf("%s/contract/%s/%s/erc1155/get-owned?walletAddress=%s",
		config.EngineCloudBaseURL,
		chainID,
		contractAddress,
		walletAddress,
	)
	println("Fetching NFTs from URL:", url)

	// Create the request using Fiber's client
	req := fiber.Get(url)
	req.Set("Authorization", "Bearer "+ws.secretKey)

	// Send the request
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
