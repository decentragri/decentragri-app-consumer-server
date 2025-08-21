package marketplaceservices

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BuyFromListingRequest represents the request to buy a token from a direct listing
type BuyFromListingRequest struct {
	ListingID string `json:"listingId"`
	Quantity  string `json:"quantity"`
	Buyer     string `json:"buyer"`
}

// BuyFromListingResponse represents the response from buying a token
type BuyFromListingResponse struct {
	Receipt json.RawMessage `json:"receipt"`
}

// CurrencyValuePerToken represents the token currency information and value
type CurrencyValuePerToken struct {
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Decimals     int    `json:"decimals"`
	Value        string `json:"value"`
	DisplayValue string `json:"displayValue"`
}

// DirectListing represents a single direct listing in the marketplace
type DirectListing struct {
	ID                      string                 `json:"id"`
	AssetContractAddress    string                 `json:"assetContractAddress"`
	TokenID                 string                 `json:"tokenId"`
	Seller                  string                 `json:"seller,omitempty"`
	PricePerToken           string                 `json:"pricePerToken"`
	CurrencyContractAddress string                 `json:"currencyContractAddress"`
	Quantity                string                 `json:"quantity"`
	IsReservedListing       bool                   `json:"isReservedListing"`
	CurrencyValuePerToken   *CurrencyValuePerToken `json:"currencyValuePerToken"`
	StartTimeInSeconds      int64                  `json:"startTimeInSeconds"`
	EndTimeInSeconds        int64                  `json:"endTimeInSeconds"`
	Status                  ListingStatus          `json:"status"`
}

type FarmPlotDirectListing struct {
	DirectListing
	Asset FarmPlotMetadata `json:"asset"`
}

// ByteArray is a custom type that marshals to JSON as an array of numbers instead of base64
type ByteArray []uint8

// MarshalJSON implements custom JSON marshaling to output as array of numbers
func (b ByteArray) MarshalJSON() ([]byte, error) {
	if b == nil {
		return []byte("null"), nil
	}

	result := make([]string, len(b))
	for i, v := range b {
		result[i] = fmt.Sprintf("%d", v)
	}
	return []byte("[" + strings.Join(result, ",") + "]"), nil
}

type FarmPlotDirectListingsWithImageByte struct {
	DirectListing
	Asset      FarmPlotMetadata `json:"asset"`
	ImageBytes ByteArray        `json:"imageBytes,omitempty"`
}

type ListingStatus string

const (
	StatusUnset     ListingStatus = "UNSET"
	StatusCreated   ListingStatus = "CREATED"
	StatusCompleted ListingStatus = "COMPLETED"
	StatusCancelled ListingStatus = "CANCELLED"
	StatusActive    ListingStatus = "ACTIVE"
	StatusExpired   ListingStatus = "EXPIRED"
)

// DirectListingsResponse represents the response from getting all direct listings
type DirectListingsResponse struct {
	Result []DirectListing `json:"result"`
}

// FarmPlotDirectListingsResponse is now just an array of listings (no wrapper)
type FarmPlotDirectListingsResponse []FarmPlotDirectListingsWithImageByte

type NFTMetadata struct {
	Name            string         `json:"name"`
	Description     string         `json:"description,omitempty"`
	Image           string         `json:"image,omitempty"`
	ExternalURL     string         `json:"external_url,omitempty"`
	BackgroundColor string         `json:"background_color,omitempty"`
	Properties      map[string]any `json:"properties,omitempty"`
	Attributes      []any          `json:"attributes,omitempty"` // Generic attributes that can contain various types
}

type FarmPlotMetadata struct {
	Name            string               `json:"name"`
	Description     string               `json:"description,omitempty"`
	Image           string               `json:"image,omitempty"`
	ExternalURL     string               `json:"external_url,omitempty"`
	BackgroundColor string               `json:"background_color,omitempty"`
	Properties      map[string]any       `json:"properties,omitempty"`
	Attributes      []FarmPlotAttributes `json:"attributes,omitempty"` // Specific for farm plot data
}

// Custom unmarshaling to handle different attribute formats
func (fpm *FarmPlotMetadata) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a temporary struct to get all the basic fields
	type Alias FarmPlotMetadata
	aux := &struct {
		*Alias
		RawAttributes []json.RawMessage `json:"attributes,omitempty"`
	}{
		Alias: (*Alias)(fpm),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try to parse attributes in different formats
	if len(aux.RawAttributes) > 0 {
		// First, try to parse as FarmPlotAttributes
		var tempStruct struct {
			*Alias
			Attributes []FarmPlotAttributes `json:"attributes,omitempty"`
		}
		tempStruct.Alias = (*Alias)(fpm)

		if err := json.Unmarshal(data, &tempStruct); err == nil && len(tempStruct.Attributes) > 0 {
			// Check if any attributes have non-empty values
			hasData := false
			for _, attr := range tempStruct.Attributes {
				if attr.ID != "" || attr.FarmName != "" || attr.Description != "" {
					hasData = true
					break
				}
			}
			if hasData {
				fpm.Attributes = tempStruct.Attributes
				return nil
			}
		}

		// If FarmPlotAttributes are empty, try to parse as standard NFT attributes
		var attrStruct struct {
			Attributes []Attribute `json:"attributes"`
		}
		attrData, _ := json.Marshal(map[string]interface{}{"attributes": aux.RawAttributes})
		if err := json.Unmarshal(attrData, &attrStruct); err == nil && len(attrStruct.Attributes) > 0 {
			// Convert standard attributes to FarmPlotAttributes
			farmPlotAttr := FarmPlotAttributes{}
			for _, attr := range attrStruct.Attributes {
				switch attr.TraitType {
				case "id":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.ID = v
					}
				case "farmName":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.FarmName = v
					}
				case "description":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.Description = v
					}
				case "cropType":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.CropType = v
					}
				case "owner":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.Owner = v
					}
				case "image":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.Image = v
					}
				case "location":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.Location = v
					}
				case "price":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.Price = v
					}
				case "createdAt":
					if v, ok := attr.Value.(string); ok {
						farmPlotAttr.CreatedAt = v
					}
				}
			}
			// Only add if we found some data
			if farmPlotAttr.ID != "" || farmPlotAttr.FarmName != "" {
				fpm.Attributes = []FarmPlotAttributes{farmPlotAttr}
				return nil
			}
		}

		// If we can't parse attributes properly, check properties field
		if fpm.Properties != nil {
			farmPlotAttr := FarmPlotAttributes{}
			if v, ok := fpm.Properties["id"].(string); ok {
				farmPlotAttr.ID = v
			}
			if v, ok := fpm.Properties["farmName"].(string); ok {
				farmPlotAttr.FarmName = v
			}
			if v, ok := fpm.Properties["description"].(string); ok {
				farmPlotAttr.Description = v
			}
			if v, ok := fpm.Properties["cropType"].(string); ok {
				farmPlotAttr.CropType = v
			}
			if v, ok := fpm.Properties["owner"].(string); ok {
				farmPlotAttr.Owner = v
			}
			if v, ok := fpm.Properties["image"].(string); ok {
				farmPlotAttr.Image = v
			}
			if v, ok := fpm.Properties["location"].(string); ok {
				farmPlotAttr.Location = v
			}
			if v, ok := fpm.Properties["price"].(string); ok {
				farmPlotAttr.Price = v
			}
			if v, ok := fpm.Properties["createdAt"].(string); ok {
				farmPlotAttr.CreatedAt = v
			}

			// Only add if we found some data
			if farmPlotAttr.ID != "" || farmPlotAttr.FarmName != "" {
				fpm.Attributes = []FarmPlotAttributes{farmPlotAttr}
			}
		}
	}

	return nil
}

type Attribute struct {
	TraitType string `json:"trait_type"`
	Value     any    `json:"value"`
}

type FarmPlotAttributes struct {
	ID          string      `json:"id"`
	Price       string      `json:"price"`
	FarmName    string      `json:"farmName"`
	Description string      `json:"description"`
	CropType    string      `json:"cropType"`
	Owner       string      `json:"owner"`
	Image       string      `json:"image"`
	Location    string      `json:"location"`
	Coordinates Coordinates `json:"coordinates"`
	CreatedAt   string      `json:"createdAt"`
}

type Coordinates struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

// ListingStatus supports both string and number JSON values
func (ls *ListingStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*ls = ListingStatus(s)
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		// Map known numbers to string status
		switch n {
		case 0:
			*ls = StatusUnset
		case 1:
			*ls = StatusCreated
		case 2:
			*ls = StatusCompleted
		case 3:
			*ls = StatusCancelled
		case 4:
			*ls = StatusActive
		case 5:
			*ls = StatusExpired
		default:
			*ls = ListingStatus(fmt.Sprintf("%d", n))
		}
		return nil
	}
	return fmt.Errorf("invalid ListingStatus: %s", string(data))
}
