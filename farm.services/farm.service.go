package farmservices

import (
	"fmt"
	"log"
	"sync"
	"time"

	"decentragri-app-cx-server/cache"
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

		formattedUpdatedAt := ""
		if !updatedAt.IsZero() {
			formattedUpdatedAt = updatedAt.Format("January 2, 2006")
		} else {
			log.Printf("Zero time detected for farm updatedAt, rawUpdatedAt: %v", rawUpdatedAt)
			formattedUpdatedAt = "Date unavailable"
		}

		rawCreatedAt, _ := record.Get("createdAt")
		createdAt := parseDate(rawCreatedAt)

		formattedCreatedAt := ""
		if !createdAt.IsZero() {
			formattedCreatedAt = createdAt.Format("January 2, 2006")
		} else {
			log.Printf("Zero time detected for farm createdAt, rawCreatedAt: %v", rawCreatedAt)
			formattedCreatedAt = "Date unavailable"
		}

		// Fetch image bytes
		imageURL, _ := record.Get("image")
		imageBytes := ByteArray{}
		if s, ok := imageURL.(string); ok && s != "" {
			log.Printf("Fetching image for farm: %s", getString(record, "farmName"))

			// Convert IPFS URL to HTTP gateway URL if needed
			httpURL := marketplaceservices.BuildIpfsUri(s)

			img, err := marketplaceservices.FetchImageBytes(httpURL)
			if err != nil {
				log.Printf("Error fetching image bytes for farm %s: %v", getString(record, "farmName"), err)
			} else {
				imageBytes = ByteArray(img)
				log.Printf("Successfully fetched %d bytes for farm: %s", len(imageBytes), getString(record, "farmName"))
			}
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
		// Try multiple date formats that might be returned from the database
		formats := []string{
			time.RFC3339,               // 2006-01-02T15:04:05Z07:00
			time.RFC3339Nano,           // 2006-01-02T15:04:05.999999999Z07:00
			"2006-01-02T15:04:05.000Z", // ISO 8601 with milliseconds
			"2006-01-02T15:04:05Z",     // ISO 8601 without milliseconds
			"2006-01-02 15:04:05",      // Standard format without timezone
			"2006-01-02T15:04:05.999Z", // ISO 8601 with variable milliseconds
			"2006-01-02T15:04:05.99Z",  // ISO 8601 with 2 digit milliseconds
			"2006-01-02T15:04:05.9Z",   // ISO 8601 with 1 digit milliseconds
		}

		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}

		// If all parsing attempts fail, return zero time
		return time.Time{}
	case int64:
		return time.Unix(v, 0)
	case float64:
		// Handle cases where timestamp might be returned as float64
		return time.Unix(int64(v), 0)
	case nil:
		return time.Time{}
	default:
		return time.Time{}
	}
} // getString safely gets a string from record
func getString(record *neo4j.Record, key string) string {
	val, _ := record.Get(key)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetFarmScans fetches recent farm scans with pagination (plant scans and soil readings) - OPTIMIZED VERSION
func GetFarmScans(farmName string, page, limit int) (*FarmScanResult, error) {
	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Set default pagination values
	if limit <= 0 {
		limit = 10 // Default to 10 items per page
	}
	if page <= 0 {
		page = 1 // Default to first page
	}

	// Check cache first - cache key includes pagination params
	cacheKey := fmt.Sprintf("farm_scans:%s:page_%d:limit_%d", farmName, page, limit)
	var cachedResult FarmScanResult
	if cache.Exists(cacheKey) {
		err := cache.Get(cacheKey, &cachedResult)
		if err == nil {
			return &cachedResult, nil
		}
	}

	// Query for plant scans with pagination - using the correct 'date' field
	plantScansCypher := `
		MATCH (f:Farm {farmName: $farmName})-[:HAS_PLANT_SCAN]->(ps:PlantScan)
		WITH ps ORDER BY COALESCE(ps.date, ps.createdAt, ps.created_at, ps.timestamp, '1970-01-01T00:00:00Z') DESC
		RETURN ps.cropType as cropType,
			   ps.note as note,
			   ps.date as date,
			   ps.createdAt as createdAt,
			   ps.created_at as created_at,
			   ps.timestamp as timestamp,
			   ps.id as id,
			   ps.interpretation as interpretation,
			   ps.imageUri as imageUri,
			   properties(ps) as allProperties
		SKIP $offset LIMIT $limit
	`

	// Query for soil readings with pagination - corrected relationship path
	soilReadingsCypher := `
		MATCH (f:Farm {farmName: $farmName})-[:HAS_SENSOR]->(s:Sensor)-[:HAS_READING]->(r:Reading)
		OPTIONAL MATCH (r)-[:INTERPRETED_AS]->(i:Interpretation)
		WITH r, i ORDER BY r.createdAt DESC
		RETURN r.fertility as fertility,
			   r.moisture as moisture,
			   r.ph as ph,
			   r.temperature as temperature,
			   r.sunlight as sunlight,
			   r.humidity as humidity,
			   r.farmName as farmName,
			   r.cropType as cropType,
			   r.sensorId as sensorId,
			   r.id as id,
			   r.createdAt as createdAt,
			   r.submittedAt as submittedAt,
			   i.value as interpretation
		SKIP $offset LIMIT $limit
	`

	// Count queries for pagination - simplified to only use farmName
	plantScansCountCypher := `
		MATCH (f:Farm {farmName: $farmName})-[:HAS_PLANT_SCAN]->(ps:PlantScan)
		RETURN COUNT(ps) as total
	`

	soilReadingsCountCypher := `
		MATCH (f:Farm {farmName: $farmName})-[:HAS_SENSOR]->(s:Sensor)-[:HAS_READING]->(r:Reading)
		RETURN COUNT(r) as total
	`

	params := map[string]interface{}{
		"farmName": farmName,
		"offset":   offset,
		"limit":    limit,
	}

	// OPTIMIZATION: Execute all 4 database queries concurrently using goroutines
	var wg sync.WaitGroup
	var plantScanRecords, soilReadingRecords, plantCountRecords, soilCountRecords []*neo4j.Record
	var plantErr, soilErr, plantCountErr, soilCountErr error

	wg.Add(4)

	// Concurrent query execution
	go func() {
		defer wg.Done()
		plantScanRecords, plantErr = memgraph.ExecuteRead(plantScansCypher, params)
	}()

	go func() {
		defer wg.Done()
		soilReadingRecords, soilErr = memgraph.ExecuteRead(soilReadingsCypher, params)
	}()

	go func() {
		defer wg.Done()
		plantCountRecords, plantCountErr = memgraph.ExecuteRead(plantScansCountCypher, map[string]interface{}{
			"farmName": farmName,
		})
	}()

	go func() {
		defer wg.Done()
		soilCountRecords, soilCountErr = memgraph.ExecuteRead(soilReadingsCountCypher, map[string]interface{}{
			"farmName": farmName,
		})
	}()

	wg.Wait()

	// Check for errors
	if plantErr != nil {
		return nil, fmt.Errorf("failed to fetch plant scans: %w", plantErr)
	}
	if soilErr != nil {
		return nil, fmt.Errorf("failed to fetch soil readings: %w", soilErr)
	}
	if plantCountErr != nil {
		return nil, fmt.Errorf("failed to get plant scans count: %w", plantCountErr)
	}
	if soilCountErr != nil {
		return nil, fmt.Errorf("failed to get soil readings count: %w", soilCountErr)
	}

	// OPTIMIZATION: Process plant scans with concurrent image fetching
	plantScans := make([]PlantScanResult, len(plantScanRecords))
	if len(plantScanRecords) > 0 {
		var imageWg sync.WaitGroup
		imageWg.Add(len(plantScanRecords))

		for i, record := range plantScanRecords {
			go func(index int, rec *neo4j.Record) {
				defer imageWg.Done()

				// Process date fields
				rawDate, dateExists := rec.Get("date")
				rawCreatedAt, _ := rec.Get("createdAt")

				var actualDateValue interface{}
				if dateExists && rawDate != nil {
					actualDateValue = rawDate
				} else if rawCreatedAt != nil {
					actualDateValue = rawCreatedAt
				} else {
					actualDateValue = nil
				}

				createdAt := parseDate(actualDateValue)
				formattedCreatedAt := ""
				if !createdAt.IsZero() {
					formattedCreatedAt = createdAt.Format("January 2, 2006 - 3:04pm")
				} else {
					formattedCreatedAt = "Date unavailable"
				}

				// Fetch image concurrently (this will use cache if available)
				imageURI, _ := rec.Get("imageUri")
				imageBytes := ByteArray{}
				if s, ok := imageURI.(string); ok && s != "" {
					httpURL := marketplaceservices.BuildIpfsUri(s)
					if img, err := marketplaceservices.FetchImageBytes(httpURL); err == nil {
						imageBytes = ByteArray(img)
					}
				}

				plantScans[index] = PlantScanResult{
					CropType:           getString(rec, "cropType"),
					Note:               getString(rec, "note"),
					CreatedAt:          createdAt,
					FormattedCreatedAt: formattedCreatedAt,
					ID:                 getString(rec, "id"),
					Interpretation:     parsePlantScanInterpretation(rec, "interpretation"),
					ImageURI:           getString(rec, "imageUri"),
					ImageBytes:         imageBytes,
				}
			}(i, record)
		}

		imageWg.Wait()
	}

	// Process soil readings (no images, so sequential processing is fine)
	soilReadings := make([]SensorReadingsWithInterpretation, 0, len(soilReadingRecords))
	for _, record := range soilReadingRecords {
		rawCreatedAt, _ := record.Get("createdAt")
		createdAt := parseDate(rawCreatedAt)

		formattedCreatedAt := ""
		if !createdAt.IsZero() {
			formattedCreatedAt = createdAt.Format("January 2, 2006 - 3:04pm")
		} else {
			formattedCreatedAt = "Date unavailable"
		}

		rawSubmittedAt, _ := record.Get("submittedAt")
		submittedAt := parseDate(rawSubmittedAt)

		formattedSubmittedAt := ""
		if !submittedAt.IsZero() {
			formattedSubmittedAt = submittedAt.Format("January 2, 2006 - 3:04pm")
		} else {
			formattedSubmittedAt = "Date unavailable"
		}

		// Parse sensor reading values
		fertility, _ := getFloat64(record, "fertility")
		moisture, _ := getFloat64(record, "moisture")
		ph, _ := getFloat64(record, "ph")
		temperature, _ := getFloat64(record, "temperature")
		sunlight, _ := getFloat64(record, "sunlight")
		humidity, _ := getFloat64(record, "humidity")

		// Parse interpretation from the connected Interpretation node
		interpretation := parseInterpretation(record, "interpretation")

		soilReading := SensorReadingsWithInterpretation{
			SensorReadings: SensorReadings{
				Fertility:            fertility,
				Moisture:             moisture,
				PH:                   ph,
				Temperature:          temperature,
				Sunlight:             sunlight,
				Humidity:             humidity,
				FarmName:             getString(record, "farmName"),
				CropType:             getString(record, "cropType"),
				SensorID:             getString(record, "sensorId"),
				ID:                   getString(record, "id"),
				CreatedAt:            createdAt,
				SubmittedAt:          submittedAt,
				FormattedCreatedAt:   formattedCreatedAt,
				FormattedSubmittedAt: formattedSubmittedAt,
			},
			Interpretation: interpretation,
		}
		soilReadings = append(soilReadings, soilReading)
	}

	// Calculate pagination info
	plantTotal := 0
	if len(plantCountRecords) > 0 {
		if total, ok := plantCountRecords[0].Get("total"); ok {
			if t, ok := total.(int64); ok {
				plantTotal = int(t)
			}
		}
	}

	soilTotal := 0
	if len(soilCountRecords) > 0 {
		if total, ok := soilCountRecords[0].Get("total"); ok {
			if t, ok := total.(int64); ok {
				soilTotal = int(t)
			}
		}
	}

	// For simplicity, we'll use the max of both totals for overall pagination
	total := plantTotal
	if soilTotal > total {
		total = soilTotal
	}

	totalPages := (total + limit - 1) / limit // Ceiling division
	hasNext := page < totalPages
	hasPrevious := page > 1

	pagination := PaginationInfo{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}

	result := &FarmScanResult{
		PlantScans:   plantScans,
		SoilReadings: soilReadings,
		Pagination:   pagination,
	}

	// Cache the result for 5 minutes to speed up subsequent requests
	cache.Set(cacheKey, *result, 5*time.Minute)

	return result, nil
}

// WarmFarmScansCache pre-loads farm scans data into cache for faster subsequent requests
// This can be called periodically or after data updates to ensure cache is warm
func WarmFarmScansCache(farmName string) error {
	// Warm cache for common page/limit combinations
	commonCombinations := []struct{ page, limit int }{
		{1, 10}, // Most common: first page, 10 items
		{1, 20}, // First page, 20 items
		{2, 10}, // Second page, 10 items
	}

	for _, combo := range commonCombinations {
		_, err := GetFarmScans(farmName, combo.page, combo.limit)
		if err != nil {
			return fmt.Errorf("failed to warm cache for page %d, limit %d: %w", combo.page, combo.limit, err)
		}
	}

	return nil
}

// getFloat64 safely gets a float64 from record
func getFloat64(record *neo4j.Record, key string) (float64, bool) {
	val, exists := record.Get(key)
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

// parseInterpretation safely parses interpretation data from the database
func parseInterpretation(record *neo4j.Record, key string) Interpretation {
	// Default interpretation values
	defaultInterpretation := Interpretation{
		Evaluation:  "Not analyzed",
		Fertility:   "No data available",
		Moisture:    "No data available",
		PH:          "No data available",
		Temperature: "No data available",
		Sunlight:    "No data available",
		Humidity:    "No data available",
	}

	val, exists := record.Get(key)
	if !exists {
		return defaultInterpretation
	}

	// If val is nil, return default
	if val == nil {
		return defaultInterpretation
	}

	// Try to parse as map[string]interface{} (which is how Neo4j returns objects)
	if interpretationMap, ok := val.(map[string]interface{}); ok {
		interpretation := Interpretation{}

		// Extract each field with safe type assertion
		if evaluation, ok := interpretationMap["evaluation"].(string); ok {
			interpretation.Evaluation = evaluation
		} else {
			interpretation.Evaluation = defaultInterpretation.Evaluation
		}

		if fertility, ok := interpretationMap["fertility"].(string); ok {
			interpretation.Fertility = fertility
		} else {
			interpretation.Fertility = defaultInterpretation.Fertility
		}

		if moisture, ok := interpretationMap["moisture"].(string); ok {
			interpretation.Moisture = moisture
		} else {
			interpretation.Moisture = defaultInterpretation.Moisture
		}

		if ph, ok := interpretationMap["ph"].(string); ok {
			interpretation.PH = ph
		} else {
			interpretation.PH = defaultInterpretation.PH
		}

		if temperature, ok := interpretationMap["temperature"].(string); ok {
			interpretation.Temperature = temperature
		} else {
			interpretation.Temperature = defaultInterpretation.Temperature
		}

		if sunlight, ok := interpretationMap["sunlight"].(string); ok {
			interpretation.Sunlight = sunlight
		} else {
			interpretation.Sunlight = defaultInterpretation.Sunlight
		}

		if humidity, ok := interpretationMap["humidity"].(string); ok {
			interpretation.Humidity = humidity
		} else {
			interpretation.Humidity = defaultInterpretation.Humidity
		}

		if historical, ok := interpretationMap["historicalComparison"].(string); ok {
			interpretation.HistoricalComparison = historical
		}

		return interpretation
	}

	// If we can't parse it, return default
	return defaultInterpretation
}

// parsePlantScanInterpretation safely parses plant scan interpretation data from the database
func parsePlantScanInterpretation(record *neo4j.Record, key string) interface{} {
	val, exists := record.Get(key)
	if !exists {
		return ""
	}

	// If val is nil, return empty string
	if val == nil {
		return ""
	}

	// Try to parse as map[string]interface{} (which is how Neo4j returns objects)
	if interpretationMap, ok := val.(map[string]interface{}); ok {
		interpretation := ParsedInterpretation{}

		// Extract diagnosis
		if diagnosis, ok := interpretationMap["diagnosis"].(string); ok {
			interpretation.Diagnosis = diagnosis
		} else if diagnosis, ok := interpretationMap["Diagnosis"].(string); ok {
			interpretation.Diagnosis = diagnosis
		}

		// Extract reason
		if reason, ok := interpretationMap["reason"].(string); ok {
			interpretation.Reason = reason
		} else if reason, ok := interpretationMap["Reason"].(string); ok {
			interpretation.Reason = reason
		}

		// Extract recommendations - handle both string and array cases
		if recommendations, ok := interpretationMap["recommendations"].([]interface{}); ok {
			for _, rec := range recommendations {
				if recStr, ok := rec.(string); ok {
					interpretation.Recommendations = append(interpretation.Recommendations, recStr)
				}
			}
		} else if recommendations, ok := interpretationMap["Recommendations"].([]interface{}); ok {
			for _, rec := range recommendations {
				if recStr, ok := rec.(string); ok {
					interpretation.Recommendations = append(interpretation.Recommendations, recStr)
				}
			}
		} else if recommendationsStr, ok := interpretationMap["recommendations"].(string); ok {
			// If recommendations is stored as a single string, put it in an array
			interpretation.Recommendations = []string{recommendationsStr}
		} else if recommendationsStr, ok := interpretationMap["Recommendations"].(string); ok {
			interpretation.Recommendations = []string{recommendationsStr}
		}

		// Extract historical comparison
		if historical, ok := interpretationMap["historicalComparison"].(string); ok {
			interpretation.HistoricalComparison = historical
		} else if historical, ok := interpretationMap["HistoricalComparison"].(string); ok {
			interpretation.HistoricalComparison = historical
		}

		return interpretation
	}

	// If it's a string, return it as is
	if str, ok := val.(string); ok {
		return str
	}

	// For any other type, return empty string
	return ""
}
