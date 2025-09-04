package farmservices

import "time"


type FarmCoordinates struct {
    Lat float64 `json:"lat"`
    Lng float64 `json:"lng"`
}

type FarmList struct {
    Owner               string         `json:"owner"`
    FarmName            string         `json:"farmName"`
    ID                  string         `json:"id"`
    CropType            string         `json:"cropType"`
    Description         string         `json:"description"`
    Image               string         `json:"image"`
    Coordinates         FarmCoordinates `json:"coordinates"`
    UpdatedAt           time.Time      `json:"updatedAt"`
    CreatedAt           time.Time      `json:"createdAt"`
    FormattedUpdatedAt  string         `json:"formattedUpdatedAt"`
    FormattedCreatedAt  string         `json:"formattedCreatedAt"`
    ImageBytes          []byte         `json:"imageBytes"`
    Location            string         `json:"location"`
}