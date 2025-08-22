package marketplaceservices

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"decentragri-app-cx-server/config"
	tokenServices "decentragri-app-cx-server/token.services"

	"github.com/gofiber/fiber/v2"
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

	walletAddr, err := tokenServices.NewTokenService().VerifyAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}
	// Set the buyer to the authenticated wallet address
	req.Buyer = walletAddr

	// Prepare the request URL
	url := fmt.Sprintf("%s/marketplace/%s/%s/direct-listings/buy-from-listing",
		config.EngineCloudBaseURL,
		config.CHAIN,
		config.MarketPlaceContractAddress,
	)

	// Create the request using Fiber's client
	fiberReq := fiber.Post(url)
	fiberReq.Set("Content-Type", "application/json")
	fiberReq.Set("Authorization", "Bearer "+os.Getenv("SECRET_KEY"))
	fiberReq.Set("X-Backend-Wallet-Address", config.AdminWallet)
	fiberReq.JSON(req) // Set JSON body

	// Send the request
	status, body, errs := fiberReq.Bytes()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to send request: %v", errs[0])
	}

	// Check response status
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	// Parse the engine response
	var engineResp EngineResponse
	if err := json.Unmarshal(body, &engineResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// For now we'll return immediately after getting the response
	// In a production environment, you might want to implement transaction mining check here

	// Create the final response
	result := &BuyFromListingResponse{
		Message: "Purchase successful",
	}
	return result, nil
}
