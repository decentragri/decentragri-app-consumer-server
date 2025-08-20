package marketplaceservices

import (
	"fmt"
	"math/rand"
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

	fmt.Println(farmPlotListing)

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
