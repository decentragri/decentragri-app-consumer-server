package farmservices

import (
	"encoding/json"
	"time"
)

type FarmCoordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type FarmList struct {
	Owner              string          `json:"owner"`
	FarmName           string          `json:"farmName"`
	ID                 string          `json:"id"`
	CropType           string          `json:"cropType"`
	Description        string          `json:"description"`
	Image              string          `json:"image"`
	Coordinates        FarmCoordinates `json:"coordinates"`
	UpdatedAt          time.Time       `json:"updatedAt"`
	CreatedAt          time.Time       `json:"createdAt"`
	FormattedUpdatedAt string          `json:"formattedUpdatedAt"`
	FormattedCreatedAt string          `json:"formattedCreatedAt"`
	ImageBytes         ByteArray       `json:"imageBytes"`
	Location           string          `json:"location"`
}

// ParsedInterpretation represents the parsed interpretation of a plant scan result
type ParsedInterpretation struct {
	Diagnosis            string   `json:"diagnosis"`
	Reason               string   `json:"reason"`
	Recommendations      []string `json:"recommendations"`
	HistoricalComparison string   `json:"historicalComparison,omitempty"`
}

// PlantScanResult represents a plant scan with analysis
type PlantScanResult struct {
	CropType           string      `json:"cropType"`
	Note               string      `json:"note"`
	CreatedAt          time.Time   `json:"createdAt"`
	FormattedCreatedAt string      `json:"formattedCreatedAt"`
	ID                 string      `json:"id"`
	Interpretation     interface{} `json:"interpretation"` // Can be string or ParsedInterpretation
	ImageURI           string      `json:"imageUri"`
	ImageBytes         ByteArray   `json:"imageBytes"`
}

// ByteArray is a custom type that marshals as an array of numbers instead of base64
type ByteArray []byte

// MarshalJSON implements json.Marshaler interface to return byte array as numbers
func (ba ByteArray) MarshalJSON() ([]byte, error) {
	if ba == nil {
		return []byte("null"), nil
	}

	// Convert to array of numbers
	result := make([]int, len(ba))
	for i, b := range ba {
		result[i] = int(b)
	}

	// Use Go's built-in JSON marshaling for the int slice
	return json.Marshal(result)
}

// SensorReadings represents sensor data collected from agricultural sensors
type SensorReadings struct {
	Fertility            float64   `json:"fertility"`
	Moisture             float64   `json:"moisture"`
	PH                   float64   `json:"ph"`
	Temperature          float64   `json:"temperature"`
	Sunlight             float64   `json:"sunlight"`
	Humidity             float64   `json:"humidity"`
	FarmName             string    `json:"farmName"`
	CropType             string    `json:"cropType"`
	SensorID             string    `json:"sensorId"`
	ID                   string    `json:"id"`
	CreatedAt            time.Time `json:"createdAt"`
	SubmittedAt          time.Time `json:"submittedAt"`
	FormattedCreatedAt   string    `json:"formattedCreatedAt"`
	FormattedSubmittedAt string    `json:"formattedSubmittedAt"`
}

// Interpretation contains human-readable interpretations of sensor readings
type Interpretation struct {
	Evaluation           string `json:"evaluation"`
	Fertility            string `json:"fertility"`
	Moisture             string `json:"moisture"`
	PH                   string `json:"ph"`
	Temperature          string `json:"temperature"`
	Sunlight             string `json:"sunlight"`
	Humidity             string `json:"humidity"`
	HistoricalComparison string `json:"historicalComparison,omitempty"`
}

// SensorReadingsWithInterpretation extends SensorReadings to include AI-generated interpretations
type SensorReadingsWithInterpretation struct {
	SensorReadings
	Interpretation Interpretation `json:"interpretation"`
}

// FarmScanResult represents the result of farm scans with pagination
type FarmScanResult struct {
	PlantScans   []PlantScanResult                  `json:"plantScans"`
	SoilReadings []SensorReadingsWithInterpretation `json:"soilReadings"`
	Pagination   PaginationInfo                     `json:"pagination"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"totalPages"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}
