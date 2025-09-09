package marketplaceservices

import (
	"crypto/md5"
	"decentragri-app-cx-server/cache"
	"decentragri-app-cx-server/config"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetAllValidFarmPlotListings(chainID, contractAddress string) (*FarmPlotDirectListingsResponse, error) {
	if chainID == "" {
		chainID = config.CHAIN
	}

	if contractAddress == "" {
		contractAddress = config.MarketPlaceContractAddress
	}

	// Create cache key
	cacheKey := fmt.Sprintf("farm_plot_listings:%s:%s", chainID, contractAddress)

	// Try to get from cache first
	var cachedResult FarmPlotDirectListingsResponse
	if cache.Exists(cacheKey) {
		err := cache.Get(cacheKey, &cachedResult)
		if err == nil {
			return &cachedResult, nil
		}
	}

	// If not in cache or cache error, fetch from API
	// Prepare the request URL
	url := fmt.Sprintf("%s/marketplace/%s/%s/direct-listings/get-all-valid",
		config.EngineCloudBaseURL,
		chainID,
		contractAddress,
	)

	req := fiber.Get(url)
	req.Set("Authorization", "Bearer "+os.Getenv("SECRET_KEY"))
	req.Set("X-Backend-Wallet-Address", config.AdminWallet)

	// Send the request
	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return nil, fmt.Errorf("error sending request: %v", errs[0])
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", status, string(body))
	}

	// Parse the response from the API (still has "result" wrapper from the external API)
	var apiResponse struct {
		Result []FarmPlotDirectListing `json:"result"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %w", err)
	}

	// Early return if no listings
	if len(apiResponse.Result) == 0 {
		result := make(FarmPlotDirectListingsResponse, 0)
		return &result, nil
	}

	// Pre-allocate result with exact capacity
	result := make(FarmPlotDirectListingsResponse, len(apiResponse.Result))

	// Pre-filter listings that have image URIs and convert in one pass
	listingsWithImages := make([]int, 0, len(apiResponse.Result))
	for i, listing := range apiResponse.Result {
		result[i] = FarmPlotDirectListingsWithImageByte{
			DirectListing: listing.DirectListing,
			Asset:         listing.Asset,
			ImageBytes:    nil, // Will be populated below
		}

		// Check if this listing has an image URI
		for _, attr := range listing.Asset.Attributes {
			if attr.Image != "" {
				listingsWithImages = append(listingsWithImages, i)
				break
			}
		}
	}

	// Only fetch images if there are listings with image URIs
	if len(listingsWithImages) == 0 {
		return &result, nil
	}

	// Limit concurrent image fetches to prevent overwhelming the server
	const maxConcurrentFetches = 20
	semaphore := make(chan struct{}, maxConcurrentFetches)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, index := range listingsWithImages {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			listing := &result[idx]

			// Extract image URI (we already know it exists from pre-filtering)
			var imageURI string
			for _, attr := range listing.Asset.Attributes {
				if attr.Image != "" {
					imageURI = attr.Image
					break
				}
			}

			log.Printf("Processing image for listing %s", listing.ID)

			// Convert IPFS URI to HTTP URL if needed
			httpURL := BuildIpfsUri(imageURI)

			// Fetch image bytes
			imageBytes, err := FetchImageBytes(httpURL)
			if err != nil {
				// Log error but don't fail the entire request
				log.Printf("Warning: Failed to fetch image for listing %s: %v", listing.ID, err)
				return
			}

			// Thread-safe assignment of image bytes
			mu.Lock()
			listing.ImageBytes = ByteArray(imageBytes)
			mu.Unlock()
		}(index)
	}

	// Wait for all image fetches to complete
	wg.Wait()

	// Cache the result for 5 minutes
	cache.Set(cacheKey, result, 5*time.Minute)

	return &result, nil
}

func FetchImageBytes(imageURI string) ([]uint8, error) {
	if imageURI == "" {
		return nil, fmt.Errorf("image URI is empty")
	}

	// Create cache key for image
	hasher := md5.New()
	hasher.Write([]byte(imageURI))
	cacheKey := fmt.Sprintf("image:%s", hex.EncodeToString(hasher.Sum(nil)))

	// Try to get from cache first
	var cachedImage []uint8
	if cache.Exists(cacheKey) {
		err := cache.Get(cacheKey, &cachedImage)
		if err == nil && len(cachedImage) > 0 {
			return cachedImage, nil
		}
	}

	// If not in cache, fetch from URL
	req := fiber.Get(imageURI)
	status, resp, errs := req.Bytes()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to fetch image: %w", errs[0])
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status %d", status)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}

	// Cache the image for 1 hour
	cache.Set(cacheKey, resp, 1*time.Hour)

	return resp, nil
}

func BuildIpfsUri(ipfsURI string) string {
	clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		// Fallback to the new client ID if environment variable is not set
		clientID = "758a938bc85320ceb23c40418e01618a"
	}

	// Check if this is already an HTTPS URL with ipfscdn.io pattern
	if strings.HasPrefix(ipfsURI, "https://") && strings.Contains(ipfsURI, ".ipfscdn.io/ipfs/") {
		// Extract the existing client ID (everything between https:// and .ipfscdn.io)
		start := len("https://")
		end := strings.Index(ipfsURI, ".ipfscdn.io/ipfs/")
		if end > start {
			existingClientID := ipfsURI[start:end]
			// Replace the existing client ID with the new one
			updatedURL := strings.Replace(ipfsURI, existingClientID+".ipfscdn.io", clientID+".ipfscdn.io", 1)
			return updatedURL
		}
	}

	// Handle regular HTTP/HTTPS URLs that don't match the ipfscdn pattern
	if strings.HasPrefix(ipfsURI, "http://") || strings.HasPrefix(ipfsURI, "https://") {
		return ipfsURI
	}

	// Handle ipfs:// URIs
	if strings.HasPrefix(ipfsURI, "ipfs://") {
		ipfsHash := strings.TrimPrefix(ipfsURI, "ipfs://")
		result := fmt.Sprintf("https://%s.ipfscdn.io/ipfs/%s", clientID, ipfsHash)
		return result
	}

	// If it doesn't match any expected format, return as is
	return ipfsURI
}
