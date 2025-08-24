// Package walletservices provides wallet management functionality for the Decentragri platform.
// This package handles wallet creation, balance queries, NFT ownership verification,
// and integration with ThirdWeb Engine for blockchain operations.
//
// The service supports:
//   - Smart wallet creation using ThirdWeb Engine
//   - Native token balance queries (ETH, etc.)
//   - ERC20 token balance queries (DAGRI, etc.)
//   - NFT ownership verification
//   - Token price fetching from external APIs
//   - Multi-token portfolio management
//
// All operations require JWT authentication and automatically extract the user's
// wallet address from the provided authentication token.
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

// WalletService provides wallet management operations using ThirdWeb Engine.
// It encapsulates the ThirdWeb secret key and provides methods for wallet operations.
type WalletService struct {
	secretKey string // ThirdWeb Engine API secret key for authenticated requests
}

// NewWalletService creates a new WalletService instance with the ThirdWeb secret key.
// The secret key is loaded from the SECRET_KEY environment variable.
//
// Returns:
//   - *WalletService: A new wallet service instance ready for operations
//
// Environment Variables Required:
//   - SECRET_KEY: ThirdWeb Engine API secret key
func NewWalletService() *WalletService {
	return &WalletService{
		secretKey: os.Getenv("SECRET_KEY"),
	}
}

// CreateWallet creates a new smart wallet using ThirdWeb's backend wallet API.
// This function creates a "smart:local" type wallet which provides enhanced security
// and functionality compared to traditional EOA wallets.
//
// The function:
//  1. Extracts the user's identity from the JWT token
//  2. Calls ThirdWeb Engine to create a new smart wallet
//  3. Returns the wallet creation response including the new wallet address
//
// Parameters:
//   - token: JWT authentication token containing user identity
//
// Returns:
//   - *CreateWalletResponse: Contains the new wallet address and creation status
//   - error: Any error that occurred during wallet creation
//
// Errors:
//   - Invalid or expired JWT token
//   - ThirdWeb Engine API failures
//   - Network connectivity issues
//   - Malformed API responses
func (ws *WalletService) CreateWallet(token string) (*CreateWalletResponse, error) {
	// Extract and validate the user identity from the JWT token
	tokenService := tokenServices.NewTokenService()
	username, err := tokenService.VerifyAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Construct the ThirdWeb Engine API endpoint for wallet creation
	url := fmt.Sprintf("%s/backend-wallet/create", config.EngineCloudBaseURL)

	// Prepare the request payload for smart wallet creation
	reqBody := CreateWalletRequest{
		Type: "smart:local", // Smart wallet type for enhanced security
	}

	// Marshal the request body to JSON
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %v", err)
	}

	// Create and configure the HTTP request
	req := fiber.Post(url)
	req.Set("Content-Type", "application/json")
	req.Set("Authorization", fmt.Sprintf("Bearer %s", ws.secretKey))
	req.Body(bodyBytes)

	// Execute the HTTP request
	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return nil, fmt.Errorf("error making request: %v", errs[0])
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	// Parse the response from ThirdWeb Engine
	var response CreateWalletResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	// Set the wallet address to the authenticated username for consistency
	response.WalletAddress = username

	return &response, nil
}

// GetBalance fetches the native token balance for a specific wallet on a given blockchain.
// This function queries ThirdWeb Engine to get the current native token balance
// (e.g., ETH on Ethereum, MATIC on Polygon) for the specified wallet address.
//
// The function uses the ThirdWeb Engine REST API endpoint:
// GET /backend-wallet/{chainId}/{walletAddress}/get-balance
//
// Parameters:
//   - chainID: The blockchain chain ID (e.g., "1" for Ethereum, "137" for Polygon)
//   - walletAddress: The wallet address to query the balance for
//
// Returns:
//   - BalanceResponse: Contains both display value (human-readable) and raw value (wei)
//   - error: Any error that occurred during the balance query
//
// Response Format:
//   - DisplayValue: Human-readable balance (e.g., "1.23")
//   - Value: Raw balance in smallest unit (e.g., "1230000000000000000" for wei)
//
// Errors:
//   - Invalid chain ID or wallet address
//   - Network connectivity issues
//   - ThirdWeb Engine API failures
//   - Malformed API responses
func GetBalance(chainID, walletAddress string) (BalanceResponse, error) {
	// Construct the ThirdWeb Engine API URL for balance query
	url := fmt.Sprintf("%s/backend-wallet/%s/%s/get-balance",
		config.EngineCloudBaseURL,
		chainID,
		walletAddress,
	)

	// Create and configure the HTTP request with proper authorization
	req := fiber.Get(url)
	req.Set("Authorization", "Bearer "+os.Getenv("SECRET_KEY"))

	// Execute the request and handle potential errors
	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return BalanceResponse{}, fmt.Errorf("failed to make request: %v", errs[0])
	}

	// Validate the HTTP response status
	if status < 200 || status >= 300 {
		return BalanceResponse{}, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	// Parse the JSON response to extract balance information
	var balanceResp BalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return BalanceResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return balanceResp, nil
}

// GetERC20Balance fetches ERC20 token balance for a specific wallet and contract.
// This function queries ThirdWeb Engine to get the current ERC20 token balance
// for any ERC20-compatible token (like DAGRI, USDC, etc.) on the specified blockchain.
//
// The function uses the ThirdWeb Engine REST API endpoint:
// GET /contract/{chainId}/{contractAddress}/erc20/balance-of?wallet_address={walletAddress}
//
// Parameters:
//   - chainID: The blockchain chain ID where the token contract is deployed
//   - contractAddress: The smart contract address of the ERC20 token
//   - walletAddress: The wallet address to query the balance for
//
// Returns:
//   - BalanceResponse: Contains both display value (formatted) and raw value (wei/smallest unit)
//   - error: Any error that occurred during the balance query
//
// Usage Examples:
//   - DAGRI token balance on Polygon
//   - USDC balance on Ethereum
//   - Any ERC20-compatible token balance
//
// Errors:
//   - Invalid chain ID, contract address, or wallet address
//   - Network connectivity issues
//   - ThirdWeb Engine API failures
//   - Contract interaction failures
//   - Malformed API responses
func GetERC20Balance(chainID, contractAddress, walletAddress string) (BalanceResponse, error) {
	// Construct the ThirdWeb Engine API URL for ERC20 balance query
	url := fmt.Sprintf("%s/contract/%s/%s/erc20/balance-of?wallet_address=%s",
		config.EngineCloudBaseURL,
		chainID,
		contractAddress,
		walletAddress,
	)

	// Create and configure the HTTP request with proper authorization
	req := fiber.Get(url)
	req.Set("Authorization", "Bearer "+os.Getenv("SECRET_KEY"))

	// Execute the request and handle potential network errors
	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return BalanceResponse{}, fmt.Errorf("failed to make request: %v", errs[0])
	}

	// Validate the HTTP response status
	if status < 200 || status >= 300 {
		return BalanceResponse{}, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	// Parse the nested JSON response structure from ThirdWeb Engine
	var response struct {
		Result BalanceResponse `json:"result"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return BalanceResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Result, nil
}

// GetUserBalances retrieves comprehensive token balances for an authenticated user.
// This function is the main entry point for balance queries and aggregates multiple
// token balances including native tokens and ERC20 tokens like DAGRI.
//
// The function performs the following operations:
//  1. Validates the JWT token and extracts the wallet address
//  2. Fetches native token balance for the hardcoded chain ID (137 - Polygon)
//  3. Fetches DAGRI token balance using the ERC20 contract
//  4. Fetches current token prices from CoinGecko API
//  5. Calculates USD values for all token holdings
//  6. Returns aggregated balance information
//
// Features:
//   - Hardcoded chain ID (137 for Polygon) - no client input required
//   - Automatic wallet address extraction from JWT token
//   - Multi-token support (Native MATIC + DAGRI)
//   - Real-time price data integration from CoinGecko
//   - USD value calculations for portfolio management
//   - Comprehensive error handling for each API call
//   - Token price caching for performance optimization
//
// Parameters:
//   - token: JWT authentication token containing the user's wallet address
//
// Returns:
//   - *UserBalances: Complete balance information including native and DAGRI tokens
//   - error: Any error encountered during balance fetching or token validation
//
// Chain Configuration:
//   - Chain ID: 137 (Polygon mainnet) - hardcoded for consistency
//   - Native Token: MATIC (Polygon's native token)
//   - DAGRI Contract: Configured in config.DAGRITokenAddress
//
// Price Data:
//   - Native token prices from CoinGecko API (matic-network)
//   - DAGRI token prices from configured API endpoint
//   - USD conversion calculations for portfolio valuation
//
// Errors:
//   - Invalid or expired JWT token
//   - Network connectivity issues
//   - ThirdWeb Engine API failures
//   - CoinGecko API rate limiting or failures
//   - Contract interaction failures
//   - JSON parsing errors
func (ws *WalletService) GetUserBalances(token string) (*UserBalances, error) {
	// Extract and validate the user identity from the JWT token
	tokenService := tokenServices.NewTokenService()
	username, err := tokenService.VerifyAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Use hardcoded chain ID for consistency (8453 = Base)
	chainID := config.CHAIN
	chainInt, err := strconv.Atoi(chainID)
	if err != nil {
		return nil, fmt.Errorf("invalid chain ID: %w", err)
	}

	// Fetch native token balance (ETH on Base)
	nativeBalance, err := GetBalance(chainID, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch native balance: %w", err)
	}

	// Fetch DAGRI token balance using ERC20 contract
	dagriBalance, err := GetERC20Balance(chainID, config.DAGRIContractAddress, username)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DAGRI balance: %w", err)
	}

	// Fetch current token prices for USD calculations
	nativePrice, err := GetTokenPriceUSD(chainInt, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch native token price: %w", err)
	}

	dagriPrice, err := GetTokenPriceUSD(chainInt, config.DAGRIContractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DAGRI token price: %w", err)
	}

	// Parse balance values for USD calculations
	nativeBalanceFloat, _ := strconv.ParseFloat(nativeBalance.Result.DisplayValue, 64)
	dagriBalanceFloat, _ := strconv.ParseFloat(dagriBalance.Result.DisplayValue, 64)

	// Prepare the comprehensive balance response
	return &UserBalances{
		WalletAddress: username,
		Native: TokenBalance{
			Balance:    nativeBalance.Result.DisplayValue,
			RawBalance: nativeBalance.Result.Value,
			PriceUSD:   nativePrice,
			ValueUSD:   nativeBalanceFloat * nativePrice,
		},
		DAGRI: TokenBalance{
			Balance:    dagriBalance.Result.DisplayValue,
			RawBalance: dagriBalance.Result.Value,
			PriceUSD:   dagriPrice,
			ValueUSD:   dagriBalanceFloat * dagriPrice,
		},
		LastUpdated: time.Now().Unix(),
	}, nil
}

// GetTokenPriceUSD fetches current USD price for tokens using ThirdWeb's price API.
// This function queries ThirdWeb's Insight API to get real-time token price data
// for both native tokens (ETH, MATIC, etc.) and ERC20 tokens.
//
// The function uses the ThirdWeb Insight API endpoint:
// GET /{chainId}.insight.thirdweb.com/v1/tokens/price?address={tokenAddress}
//
// Parameters:
//   - chainID: The blockchain chain ID as integer (e.g., 1 for Ethereum, 137 for Polygon)
//   - tokenAddress: The token contract address (empty string for native tokens)
//
// Returns:
//   - float64: Current USD price of the token
//   - error: Any error that occurred during price fetching
//
// Price Data:
//   - Native tokens: Use empty string or "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
//   - ERC20 tokens: Use the actual contract address
//   - Prices are updated in real-time from multiple sources
//   - Includes market cap, volume, and other trading data
//
// Errors:
//   - Invalid chain ID or token address
//   - Network connectivity issues
//   - ThirdWeb Insight API failures
//   - Rate limiting from price feeds
//   - Token not found or not supported
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

// GetOwnedNFTs fetches owned NFTs from a specific contract for an authenticated user.
// This function queries ThirdWeb Engine to retrieve all NFTs owned by the user
// from a specific ERC1155 contract, providing comprehensive ownership data.
//
// The function performs the following operations:
//  1. Validates the JWT token and extracts the wallet address
//  2. Queries ThirdWeb Engine for NFTs owned by the user
//  3. Returns detailed NFT information including metadata and quantities
//
// Features:
//   - Automatic wallet address extraction from JWT token
//   - ERC1155 multi-token standard support
//   - Comprehensive metadata retrieval
//   - Quantity ownership tracking
//   - Error handling for API failures
//
// Parameters:
//   - contractAddress: The ERC1155 contract address to query NFTs from
//   - token: JWT authentication token containing the user's wallet address
//
// Returns:
//   - NFTResponse: Contains array of owned NFTs with metadata and quantities
//   - error: Any error encountered during NFT fetching or token validation
//
// API Endpoint:
//   - GET /contract/{chainId}/{contractAddress}/erc1155/get-owned?walletAddress={walletAddress}
//
// Response Data:
//   - NFT metadata (name, description, image, attributes)
//   - Ownership quantities for each token ID
//   - Total supply information
//   - Token type classification
//
// Errors:
//   - Invalid or expired JWT token
//   - Invalid contract address
//   - Network connectivity issues
//   - ThirdWeb Engine API failures
//   - Contract interaction failures
func (ws *WalletService) GetOwnedNFTs(contractAddress, token string) (NFTResponse, error) {
	// Extract and validate the user identity from the JWT token
	tokenService := tokenServices.NewTokenService()
	username, err := tokenService.VerifyAccessToken(token)
	if err != nil {
		return NFTResponse{}, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Construct the ThirdWeb Engine API URL for NFT ownership query
	url := fmt.Sprintf("%s/contract/%s/%s/erc1155/get-owned?walletAddress=%s",
		config.EngineCloudBaseURL,
		config.CHAIN,
		contractAddress,
		username,
	)
	println("Fetching NFTs from URL:", url)

	// Create and configure the HTTP request with proper authorization
	req := fiber.Get(url)
	req.Set("Authorization", "Bearer "+ws.secretKey)

	// Execute the request and handle potential network errors
	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return NFTResponse{}, fmt.Errorf("failed to make request: %v", errs[0])
	}

	// Validate the HTTP response status
	if status < 200 || status >= 300 {
		return NFTResponse{}, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	// Parse the JSON response to extract NFT ownership data
	var nftResp NFTResponse
	if err := json.Unmarshal(body, &nftResp); err != nil {
		return NFTResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return nftResp, nil
}
