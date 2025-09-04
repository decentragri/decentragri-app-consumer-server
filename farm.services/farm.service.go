package farmservices

import (
	"fmt"
	"time"

	memgraph "decentragri-app-cx-server/db"
	marketplaceservices "decentragri-app-cx-server/marketplace.services"

	// tokenservices "decentragri-app-cx-server/token.services"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// GetFarmList fetches farms for a user, formats dates, and fetches image bytes.
func GetFarmList() ([]FarmList, error) {
	// Handle dev bypass token first
	// var username string
	// var err error

	// if token == "dev_bypass_authorized" {
	// 	fmt.Println("Dev bypass detected in farm service")
	// 	username = "0x984785A89BF95cb3d5Df4E45F670081944d8D547" // Treasury wallet for testing
	// } else {
	// 	// Standard JWT token verification
	// 	tokenService := tokenservices.NewTokenService()
	// 	username, err = tokenService.VerifyAccessToken(token)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("token verification failed: %w", err)
	// 	}
	// }

	cypher := `
        MATCH (f:Farm)
        RETURN f.id as id, 
               f.farmName as farmName, 
               f.cropType as cropType, 
               f.description as description, 
               f.createdAt as createdAt, 
               f.updatedAt as updatedAt, 
               f.coordinates as coordinates,
               f.image as image,
               f.owner as owner,
               f.location as location,
               f.lat as lat, 
               f.lng as lng
    `

	records, err := memgraph.ExecuteRead(cypher, map[string]interface{}{})
	if err != nil {
		return []FarmList{}, fmt.Errorf("database query failed: %w", err)
	}

	if len(records) == 0 {
		return []FarmList{}, nil
	}

	farms := make([]FarmList, 0, len(records))
	for _, record := range records {
		// Parse dates
		rawUpdatedAt, _ := record.Get("updatedAt")
		updatedAt := parseDate(rawUpdatedAt)
		formattedUpdatedAt := updatedAt.Format("January 2, 2006")

		rawCreatedAt, _ := record.Get("createdAt")
		createdAt := parseDate(rawCreatedAt)
		formattedCreatedAt := createdAt.Format("January 2, 2006")

		// Fetch image bytes
		imageURL, _ := record.Get("image")
		imageBytes := []byte{}
		if s, ok := imageURL.(string); ok && s != "" {
			fmt.Printf("Fetching image for farm: %s, URL: %s\n", getString(record, "farmName"), s)
			
			// Convert IPFS URL to HTTP gateway URL if needed
			httpURL := marketplaceservices.BuildIpfsUri(s)
			fmt.Printf("Converted URL: %s\n", httpURL)
			
			img, err := marketplaceservices.FetchImageBytes(httpURL)
			if err != nil {
				fmt.Printf("Error fetching image bytes for URL %s: %v\n", httpURL, err)
			} else {
				imageBytes = img
				fmt.Printf("Successfully fetched %d bytes for URL: %s\n", len(imageBytes), httpURL)
			}
		} else {
			fmt.Printf("No image URL found for farm: %s\n", getString(record, "farmName"))
		}

		// Parse coordinates
		coords := FarmCoordinates{}
		if c, ok := record.Get("coordinates"); ok {
			if m, ok := c.(map[string]interface{}); ok {
				coords.Lat, _ = m["lat"].(float64)
				coords.Lng, _ = m["lng"].(float64)
			}
		}

		farm := FarmList{
			Owner:              getString(record, "owner"),
			FarmName:           getString(record, "farmName"),
			ID:                 getString(record, "id"),
			CropType:           getString(record, "cropType"),
			Description:        getString(record, "description"),
			Image:              getString(record, "image"),
			Coordinates:        coords,
			UpdatedAt:          updatedAt,
			CreatedAt:          createdAt,
			FormattedUpdatedAt: formattedUpdatedAt,
			FormattedCreatedAt: formattedCreatedAt,
			ImageBytes:         imageBytes,
			Location:           getString(record, "location"),
		}
		farms = append(farms, farm)
	}

	return farms, nil
}

// parseDate tries to convert interface{} to time.Time
func parseDate(val interface{}) time.Time {
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		t, _ := time.Parse(time.RFC3339, v)
		return t
	case int64:
		return time.Unix(v, 0)
	default:
		return time.Time{}
	}
}

// getString safely gets a string from record
func getString(record *neo4j.Record, key string) string {
	val, _ := record.Get(key)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}
