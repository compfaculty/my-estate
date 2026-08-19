package domain

import "time"

type SearchFilter struct {
	City         string
	Neighborhood string
	Street       string
	Query        string
	MinPrice     *float64
	MaxPrice     *float64
	MinArea      *float64
	MaxArea      *float64
	MinRooms     *float64
	MaxRooms     *float64
	MinFloor     *int
	MaxFloor     *int
	PropertyType PropertyType
	Source       string
	ListingAge   *time.Duration
	Elevator     *bool
	Parking      *bool
	Balcony      *bool
	SafeRoom     *bool
	MinLat       *float64
	MinLng       *float64
	MaxLat       *float64
	MaxLng       *float64
	Lat          *float64
	Lng          *float64
	RadiusM      *float64
	Limit        int
	Offset       int
}

func (f SearchFilter) WithDefaults() SearchFilter {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return f
}

type PropertySearchResult struct {
	Property
	AskingPrice      *float64
	Currency         string
	PricePerM2       *float64
	Source           string
	SourceURL        string
	FirstSeenAt      time.Time
	PhotoURL         string
	ListingCount     int
	PriceUAH         *float64
	PriceUSD         *float64
	PrimaryListingID string
}
