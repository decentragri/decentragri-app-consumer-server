package portfolioservices

import (
	"crypto/md5"
	"decentragri-app-cx-server/cache"
	"decentragri-app-cx-server/config"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	tokenServices "decentragri-app-cx-server/token.services"
	walletServices "decentragri-app-cx-server/wallet.services"

	"github.com/gofiber/fiber/v2"
)

type ByteArray []uint8

type PortfolioSummary struct {
	FarmPlotNFTCount int `json:"farmPlotNFTCount"`
}

type NFTItemWithImageBytes struct {
	Metadata      walletServices.NFTMetadata `json:"metadata"`
	Owner         string                     `json:"owner"`
	Type          string                     `json:"type"`
	Supply        string                     `json:"supply"`
	QuantityOwned string                     `json:"quantityOwned"`
	ImageBytes    ByteArray                  `json:"imageBytes,omitempty"`
}

type EntirePortfolio struct {
	FarmPlotNFTs []NFTItemWithImageBytes `json:"farmPlotNFTs"`
}

func GetPortFolioSummary(token string) (PortfolioSummary, error) {
	var username string
	var err error

	// Check if this is a dev bypass token
	if token == "dev_bypass_authorized" {
		fmt.Println("Dev bypass detected in portfolio service")
		username = "0x984785A89BF95cb3d5Df4E45F670081944d8D547" // Use dev wallet address
	} else {
		// Normal token verification
		username, err = tokenServices.NewTokenService().VerifyAccessToken(token)
		if err != nil {
			return PortfolioSummary{}, err
		}
	}

	// Create cache key for portfolio summary
	cacheKey := fmt.Sprintf("portfolio:%s", username)

	// Try to get from cache first
	var cachedSummary PortfolioSummary
	if cache.Exists(cacheKey) {
		err := cache.Get(cacheKey, &cachedSummary)
		if err == nil {
			return cachedSummary, nil
		}
	}

	walletService := walletServices.NewWalletService()
	farmPlotNFTs, err := walletService.GetOwnedNFTs(config.FarmPlotContractAddress, token)
	if err != nil {
		return PortfolioSummary{}, err
	}

	farmPlotNFTCount := len(farmPlotNFTs.Result)

	summary := PortfolioSummary{
		FarmPlotNFTCount: farmPlotNFTCount,
	}

	// Cache the portfolio summary for 3 minutes
	cache.Set(cacheKey, summary, 3*time.Minute)

	return summary, nil
}

func GetEntirePortfolio(token string) (EntirePortfolio, error) {
	var username string
	var err error

	// Check if this is a dev bypass token
	if token == "dev_bypass_authorized" {
		fmt.Println("Dev bypass detected in portfolio service")
		username = "0x984785A89BF95cb3d5Df4E45F670081944d8D547" // Use dev wallet address
	} else {
		// Normal token verification
		username, err = tokenServices.NewTokenService().VerifyAccessToken(token)
		if err != nil {
			return EntirePortfolio{}, err
		}
	}

	// Create cache key for entire portfolio
	cacheKey := fmt.Sprintf("entire_portfolio:%s", username)

	// Try to get from cache first
	var cachedPortfolio EntirePortfolio
	if cache.Exists(cacheKey) {
		err := cache.Get(cacheKey, &cachedPortfolio)
		if err == nil {
			return cachedPortfolio, nil
		}
	}

	walletService := walletServices.NewWalletService()
	farmPlotNFTs, err := walletService.GetOwnedNFTs(config.FarmPlotContractAddress, token)
	if err != nil {
		return EntirePortfolio{}, err
	}

	// Convert NFTResponse to NFTItemWithImageBytes and fetch images
	nftsWithImages, err := ConvertNFTsWithImages(farmPlotNFTs.Result)
	if err != nil {
		return EntirePortfolio{}, err
	}

	portfolio := EntirePortfolio{
		FarmPlotNFTs: nftsWithImages,
	}

	// Cache the entire portfolio for 3 minutes
	cache.Set(cacheKey, portfolio, 3*time.Minute)

	return portfolio, nil
}

// ConvertNFTsWithImages converts NFTItems to NFTItemWithImageBytes and fetches images
func ConvertNFTsWithImages(nftItems []walletServices.NFTItem) ([]NFTItemWithImageBytes, error) {
	result := make([]NFTItemWithImageBytes, len(nftItems))

	// Pre-filter NFTs that have image URIs
	nftsWithImages := make([]int, 0, len(nftItems))

	for i, item := range nftItems {
		result[i] = NFTItemWithImageBytes{
			Metadata:      item.Metadata,
			Owner:         item.Owner,
			Type:          item.Type,
			Supply:        item.Supply,
			QuantityOwned: item.QuantityOwned,
			ImageBytes:    nil, // Will be populated below
		}

		// Check if this NFT has an image URI in attributes
		for _, attr := range item.Metadata.Attributes {
			if attr.TraitType == "image" && attr.Value != "" {
				nftsWithImages = append(nftsWithImages, i)
				break
			}
		}

		// Also check the URI field for image
		if item.Metadata.URI != "" {
			// Check if we haven't already added this item
			found := false
			for _, idx := range nftsWithImages {
				if idx == i {
					found = true
					break
				}
			}
			if !found {
				nftsWithImages = append(nftsWithImages, i)
			}
		}
	}

	// Only fetch images if there are NFTs with image URIs
	if len(nftsWithImages) == 0 {
		return result, nil
	}

	// Limit concurrent image fetches
	const maxConcurrentFetches = 20
	semaphore := make(chan struct{}, maxConcurrentFetches)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, index := range nftsWithImages {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			nftItem := &result[idx]

			// Extract image URI
			var imageURI string

			// First check attributes for image
			for _, attr := range nftItem.Metadata.Attributes {
				if attr.TraitType == "image" && attr.Value != "" {
					imageURI = attr.Value
					break
				}
			}

			// If no image in attributes, use URI
			if imageURI == "" && nftItem.Metadata.URI != "" {
				imageURI = nftItem.Metadata.URI
			}

			if imageURI == "" {
				return
			}

			fmt.Printf("[DEBUG] Original image URI from NFT metadata: %s\n", imageURI)

			// Convert IPFS URI to HTTP URL if needed
			httpURL := BuildIpfsUri(imageURI)
			fmt.Printf("[DEBUG] HTTP URL after BuildIpfsUri: %s\n", httpURL)

			// Fetch image bytes
			imageBytes, err := FetchImageBytes(httpURL)
			if err != nil {
				fmt.Printf("Warning: Failed to fetch image for NFT %s: %v\n", nftItem.Metadata.ID, err)
				return
			}

			// Thread-safe assignment of image bytes
			mu.Lock()
			nftItem.ImageBytes = ByteArray(imageBytes)
			mu.Unlock()
		}(index)
	}

	// Wait for all image fetches to complete
	wg.Wait()

	return result, nil
}

// FetchImageBytes fetches image data from a URL and returns it as bytes
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

// BuildIpfsUri converts IPFS URIs to HTTP URLs
func BuildIpfsUri(ipfsURI string) string {
	fmt.Printf("[DEBUG] BuildIpfsUri input: %s\n", ipfsURI)

	if ipfsURI == "" {
		return ""
	}

	// If already an HTTP/HTTPS URL, return as-is
	if strings.HasPrefix(ipfsURI, "http://") || strings.HasPrefix(ipfsURI, "https://") {
		return ipfsURI
	}

	// Convert IPFS URIs to HTTP gateway URLs
	if strings.HasPrefix(ipfsURI, "ipfs://") {
		hash := strings.TrimPrefix(ipfsURI, "ipfs://")
		return fmt.Sprintf("https://ipfs.io/ipfs/%s", hash)
	}

	// If it's just a hash without protocol
	if !strings.Contains(ipfsURI, "://") && len(ipfsURI) == 46 && strings.HasPrefix(ipfsURI, "Qm") {
		return fmt.Sprintf("https://ipfs.io/ipfs/%s", ipfsURI)
	}

	// Default: assume it's already a proper URL
	return ipfsURI
}
