package walletservices

// TokenBalance represents the balance and price information for a token
type TokenBalance struct {
	Balance    string  `json:"balance"`    // Display value of the balance
	RawBalance string  `json:"rawBalance"` // Raw value of the balance
	PriceUSD   float64 `json:"priceUSD"`   // Current price in USD
	ValueUSD   float64 `json:"valueUSD"`   // Total value in USD (balance * price)
}

// UserBalances represents comprehensive balance information for a user
type UserBalances struct {
	WalletAddress string       `json:"walletAddress"`
	Native        TokenBalance `json:"native"`      // Native token (ETH) balance and price
	DAGRI         TokenBalance `json:"dagri"`       // DAGRI token balance (no price yet)
	LastUpdated   int64        `json:"lastUpdated"` // Unix timestamp of last update
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
	Type          string      `json:"type"`          // "ERC1155", "ERC721", or "metaplex"
	Supply        string      `json:"supply"`        // Total supply of the NFT
	QuantityOwned string      `json:"quantityOwned"` // Quantity owned by the user
}

// NFTAttribute represents an NFT attribute
type NFTAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

// NFTMetadata represents the metadata of an NFT
type NFTMetadata struct {
	ID          string         `json:"id"`
	URI         string         `json:"uri"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ExternalURL string         `json:"external_url"`
	Attributes  []NFTAttribute `json:"attributes"`
}

// CreateWalletRequest represents the request to create a new wallet
type CreateWalletRequest struct {
	Type string `json:"type"`
}

// CreateWalletResponse represents the response from wallet creation
type CreateWalletResponse struct {
	WalletAddress string `json:"walletAddress"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}
