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
	Metadata NFTMetadata `json:"metadata"`
	Owner    string      `json:"owner"`
	Type     string      `json:"type"` // "ERC1155", "ERC721", or "metaplex"
}

// NFTMetadata represents the metadata of an NFT
type NFTMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
