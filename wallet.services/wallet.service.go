package walletservices

import (
	"decentragri-app-cx-server/config"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	tokenServices "decentragri-app-cx-server/token.services"

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

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}
	req := fiber.Post(url)
	req.Set("Content-Type", "application/json")
	req.Set("Authorization", fmt.Sprintf("Bearer %s", ws.secretKey))
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
	url := fmt.Sprintf("%s/backend-wallet/%s/%s/get-balance",
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

// GetERC20Balance fetches ERC20 token balance using Fiber client
func GetERC20Balance(chainID, contractAddress, walletAddress string) (BalanceResponse, error) {
	url := fmt.Sprintf("%s/contract/%s/%s/erc20/balance-of?wallet_address=%s",
		config.EngineCloudBaseURL,
		chainID,
		contractAddress,
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
func (ws *WalletService) GetUserBalances(token string) (*UserBalances, error) {
	tokenService := tokenServices.NewTokenService()
	username, err := tokenService.VerifyAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}
	chainID := config.CHAIN // hardcoded chain ID from config
	chainInt, err := strconv.Atoi(chainID)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	// Get native token balance
	nativeBalance, err := GetBalance(chainID, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch native balance: %w", err)
	}

	// Get native token (ETH) price
	nativePrice, err := GetTokenPriceUSD(chainInt, "")
	if err != nil {
		nativePrice = 0 // Set price to 0 if unable to fetch
	}

	// Calculate USD value for native token using display value
	nativeBalanceFloat, _ := strconv.ParseFloat(nativeBalance.Result.DisplayValue, 64)
	nativeValueUSD := nativeBalanceFloat * nativePrice

	// Get DAGRI token balance
	dagriBalance, err := GetERC20Balance(chainID, config.DAGRIContractAddress, username)
	if err != nil {
		// If DAGRI balance fetch fails, set to zero but don't fail the entire request
		dagriBalance = BalanceResponse{
			Result: struct {
				DisplayValue string `json:"displayValue"`
				Value        string `json:"value"`
			}{
				DisplayValue: "0",
				Value:        "0",
			},
		}
	}

	// Calculate USD value for DAGRI token using display value
	// For now, DAGRI price is 0, but the structure is ready for when price is available
	dagriBalanceFloat, _ := strconv.ParseFloat(dagriBalance.Result.DisplayValue, 64)
	dagriPriceUSD := 0.0 // Set to 0 for now since DAGRI is not listed yet
	dagriValueUSD := dagriBalanceFloat * dagriPriceUSD

	// Create response
	result := &UserBalances{
		WalletAddress: username,
		Native: TokenBalance{
			Balance:    nativeBalance.Result.DisplayValue,
			RawBalance: nativeBalance.Result.Value,
			PriceUSD:   nativePrice,
			ValueUSD:   nativeValueUSD,
		},
		DAGRI: TokenBalance{
			Balance:    dagriBalance.Result.DisplayValue,
			RawBalance: dagriBalance.Result.Value,
			PriceUSD:   dagriPriceUSD,
			ValueUSD:   dagriValueUSD,
		},
		LastUpdated: time.Now().Unix(),
	}

	return result, nil
}

// GetOwnedNFTs fetches owned NFTs from a specific contract using Fiber client
func (ws *WalletService) GetOwnedNFTs(contractAddress, token string) (NFTResponse, error) {
	tokenService := tokenServices.NewTokenService()
	username, err := tokenService.VerifyAccessToken(token)
	if err != nil {
		return NFTResponse{}, fmt.Errorf("invalid or expired token: %w", err)
	}
	url := fmt.Sprintf("%s/contract/%s/%s/erc1155/get-owned?walletAddress=%s",
		config.EngineCloudBaseURL,
		config.CHAIN,
		contractAddress,
		username,
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


