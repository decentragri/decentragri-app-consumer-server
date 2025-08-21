package marketplaceservices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"decentragri-app-cx-server/config"
	tokenServices "decentragri-app-cx-server/token.services"
)

func GetValidFarmPlotListings(token string) (*FarmPlotDirectListingsResponse, error) {
	// Check for dev bypass token first
	if token == "dev_bypass_authorized" {
		fmt.Println("Dev bypass detected in marketplace service")
	} else {
		_, err := tokenServices.NewTokenService().VerifyAccessToken(token)
		if err != nil {
			return nil, err
		}
	}

	// Use the marketplace contract address to get listings, not the farm plot contract
	farmPlotListing, err := GetAllValidFarmPlotListings(config.CHAIN, config.MarketPlaceContractAddress)
	if err != nil {
		return nil, err
	}

	// Check if there are any listings
	if farmPlotListing == nil || len(*farmPlotListing) == 0 {
		return nil, fmt.Errorf("no farm plot listings available")
	}

	// The farmPlotListing already contains ImageBytes populated by GetAllValidFarmPlotListings
	return farmPlotListing, nil
}

func FeaturedProperty(token string) (*FarmPlotDirectListingsWithImageByte, error) {
	// Check for dev bypass token first
	if token == "dev_bypass_authorized" {
		fmt.Println("Dev bypass detected in marketplace service")
	} else {
		_, err := tokenServices.NewTokenService().VerifyAccessToken(token)
		if err != nil {
			return nil, err
		}
	}

	// Use the marketplace contract address to get listings
	farmPlotListing, err := GetAllValidFarmPlotListings(config.CHAIN, config.MarketPlaceContractAddress)
	if err != nil {
		return nil, err
	}

	// Check if there are any listings
	if farmPlotListing == nil || len(*farmPlotListing) == 0 {
		return nil, fmt.Errorf("no farm plot listings available")
	}

	// Get a random listing from the array
	listings := *farmPlotListing

	// Create a new random generator with a time-based seed
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rng.Intn(len(listings))

	return &listings[randomIndex], nil
}

// BuyFromListing purchases a token from a direct listing
func BuyFromListing(token string, req *BuyFromListingRequest) (*BuyFromListingResponse, error) {
	// Check for dev bypass token first
	if token == "dev_bypass_authorized" {
		fmt.Println("Dev bypass detected in marketplace service")
	} else {
		// Verify user's token
		walletAddr, err := tokenServices.NewTokenService().VerifyAccessToken(token)
		if err != nil {
			return nil, fmt.Errorf("unauthorized: %w", err)
		}
		// Set the buyer to the authenticated wallet address
		req.Buyer = walletAddr
	}

	// Marshal the request body
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Prepare the request URL
	url := fmt.Sprintf("%s/marketplace/%s/%s/direct-listings/buy-from-listing",
		config.EngineCloudBaseURL,
		config.CHAIN,
		config.MarketPlaceContractAddress,
	)

	// Create the request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+os.Getenv("SECRET_KEY"))
	httpReq.Header.Set("X-Backend-Wallet-Address", config.AdminWallet)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse the response
	var result BuyFromListingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
