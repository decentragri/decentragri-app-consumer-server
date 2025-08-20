package walletservices

import (
	"os"
)

const (
	CHAIN             = "167009" // Swell mainnet
	DECENTRAGRI_TOKEN = "0x..."  // Replace with actual token address
	RSWETH_ADDRESS    = "0x0a6E7Ba5042B38349e437ec6Db6214AEC7B35676"
)

// WalletData represents the wallet balance and price data
type WalletData struct {
	SmartWalletAddress string `json:"smartWalletAddress"`

	// Balances
	EthBalance    string `json:"ethBalance"`
	SwellBalance  string `json:"swellBalance"`
	RsWETHBalance string `json:"rsWETHBalance"`
	DagriBalance  string `json:"dagriBalance"`
	NativeBalance string `json:"nativeBalance"`

	// Prices
	DagriPriceUSD float64 `json:"dagriPriceUSD"`
	EthPriceUSD   float64 `json:"ethPriceUSD"`
	SwellPriceUSD float64 `json:"swellPriceUSD"`
}

// BalanceResponse represents the response from thirdweb balance API
type BalanceResponse struct {
	Result struct {
		DisplayValue string `json:"displayValue"`
		Value        string `json:"value"`
	} `json:"result"`
}

// PriceResponse represents the response from thirdweb price API
type PriceResponse struct {
	Data []struct {
		PriceUSD float64 `json:"price_usd"`
	} `json:"data"`
}

// NFTResponse represents the response from thirdweb NFT API
type NFTResponse struct {
	Result []NFTItem `json:"result"`
}

// NFTItem represents a single NFT item
type NFTItem struct {
	Metadata      NFTMetadata `json:"metadata"`
	Owner         string      `json:"owner"`
	Type          string      `json:"type"` // "ERC1155", "ERC721", or "metaplex"
	Supply        string      `json:"supply"`
	QuantityOwned string      `json:"quantityOwned"`
}

// NFTMetadata represents NFT metadata
type NFTMetadata struct {
	Id          string         `json:"id"`
	Uri         string         `json:"uri"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ExternalUrl string         `json:"external_url"`
	Image       string         `json:"image,omitempty"` // Optional field
	Attributes  []NFTAttribute `json:"attributes,omitempty"`
}

// NFTAttribute represents NFT attribute
type NFTAttribute struct {
	TraitType string      `json:"trait_type"`
	Value     any `json:"value"`
}

// WalletService handles wallet operations
type WalletService struct {
	secretKey string
}

// NewWalletService creates a new wallet service instance
func NewWalletService() *WalletService {
	return &WalletService{
		secretKey: os.Getenv("SECRET_KEY"),
	}
}

// InsightService handles token price fetching
type InsightService struct {
	secretKey string
}

// NewInsightService creates a new insight service instance
func NewInsightService() *InsightService {
	return &InsightService{
		secretKey: os.Getenv("SECRET_KEY"),
	}
}
