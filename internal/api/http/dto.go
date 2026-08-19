package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/google/uuid"
)

type healthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type pageResponse struct {
	Items  []propertyCard `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type geoJSON struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type propertyCard struct {
	ID           string   `json:"id"`
	Address      string   `json:"address"`
	City         string   `json:"city"`
	Neighborhood string   `json:"neighborhood"`
	Street       string   `json:"street"`
	HouseNumber  string   `json:"house_number"`
	PropertyType string   `json:"property_type"`
	Rooms        *float64 `json:"rooms"`
	AreaM2       *float64 `json:"area_m2"`
	Floor        *int     `json:"floor"`
	AskingPrice  *float64 `json:"asking_price"`
	Currency     string   `json:"currency"`
	PricePerM2   *float64 `json:"price_per_m2"`
	PriceUAH     *float64 `json:"price_uah"`
	PriceUSD     *float64 `json:"price_usd"`
	Source       string   `json:"source"`
	SourceURL    string   `json:"source_url"`
	FirstSeenAt  string   `json:"first_seen_at"`
	PhotoURL     string   `json:"photo_url"`
	ListingCount int      `json:"listing_count"`
	Location     *geoJSON `json:"location,omitempty"`
	Elevator     *bool    `json:"elevator"`
	Parking      *bool    `json:"parking"`
	Balcony      *bool    `json:"balcony"`
	SafeRoom     *bool    `json:"safe_room"`
}

type propertyDetail struct {
	Property propertyJSON  `json:"property"`
	Listings []listingJSON `json:"listings"`
}

type propertyJSON struct {
	ID               string                   `json:"id"`
	Country          string                   `json:"country"`
	City             string                   `json:"city"`
	District         string                   `json:"district"`
	Neighborhood     string                   `json:"neighborhood"`
	Street           string                   `json:"street"`
	HouseNumber      string                   `json:"house_number"`
	Address          string                   `json:"address"`
	Location         *geoJSON                 `json:"location,omitempty"`
	PropertyType     string                   `json:"property_type"`
	Rooms            *float64                 `json:"rooms"`
	Bedrooms         *int                     `json:"bedrooms"`
	Bathrooms        *float64                 `json:"bathrooms"`
	Floor            *int                     `json:"floor"`
	TotalFloors      *int                     `json:"total_floors"`
	AreaM2           *float64                 `json:"area_m2"`
	BalconyAreaM2    *float64                 `json:"balcony_area_m2"`
	ParkingSpaces    *int                     `json:"parking_spaces"`
	ConstructionYear *int                     `json:"construction_year"`
	RenovationState  string                   `json:"renovation_state"`
	Elevator         *bool                    `json:"elevator"`
	Parking          *bool                    `json:"parking"`
	Storage          *bool                    `json:"storage"`
	Balcony          *bool                    `json:"balcony"`
	AirConditioning  *bool                    `json:"air_conditioning"`
	SafeRoom         *bool                    `json:"safe_room"`
	Accessibility    *bool                    `json:"accessibility"`
	FieldProvenance  []domain.FieldProvenance `json:"field_provenance"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
}

type listingJSON struct {
	ID              string         `json:"id"`
	PropertyID      string         `json:"property_id"`
	Source          string         `json:"source"`
	SourceListingID string         `json:"source_listing_id"`
	SourceURL       string         `json:"source_url"`
	ListingType     string         `json:"listing_type"`
	AskingPrice     *float64       `json:"asking_price"`
	Currency        string         `json:"currency"`
	PricePerM2      *float64       `json:"price_per_m2"`
	PriceUAH        *float64       `json:"price_uah"`
	PriceUSD        *float64       `json:"price_usd"`
	PricePerM2UAH   *float64       `json:"price_per_m2_uah"`
	ExchangeRate    *float64       `json:"exchange_rate"`
	FxSource        string         `json:"fx_source"`
	Description     string         `json:"description"`
	SellerName      string         `json:"seller_name"`
	SellerType      string         `json:"seller_type"`
	FirstSeenAt     string         `json:"first_seen_at"`
	LastSeenAt      string         `json:"last_seen_at"`
	RemovedAt       *string        `json:"removed_at"`
	Status          string         `json:"status"`
	Photos          []domain.Photo `json:"photos"`
	MatchConfidence *float64       `json:"match_confidence"`
}

type historyJSON struct {
	ID              string   `json:"id"`
	ListingID       string   `json:"listing_id"`
	RecordedAt      string   `json:"recorded_at"`
	Price           *float64 `json:"price"`
	Currency        string   `json:"currency"`
	PriceUAH        *float64 `json:"price_uah"`
	PriceUSD        *float64 `json:"price_usd"`
	DescriptionHash string   `json:"description_hash"`
	Status          string   `json:"status"`
	SellerType      string   `json:"seller_type"`
	PhotosHash      string   `json:"photos_hash"`
}

type sourceJSON struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Enabled            bool    `json:"enabled"`
	FetchMode          string  `json:"fetch_mode"`
	LastSuccessAt      *string `json:"last_success_at"`
	LastError          *string `json:"last_error"`
	LastCrawlAt        *string `json:"last_crawl_at"`
	LastRunStatus      string  `json:"last_run_status"`
	ListingsDiscovered int     `json:"listings_discovered"`
	ListingsUpdated    int     `json:"listings_updated"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func fmtTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := fmtTime(*t)
	return &s
}

func cardFrom(r domain.PropertySearchResult) propertyCard {
	c := propertyCard{
		ID:           r.ID.String(),
		Address:      r.AddressLine(),
		City:         r.City,
		Neighborhood: r.Neighborhood,
		Street:       r.Street,
		HouseNumber:  r.HouseNumber,
		PropertyType: string(r.PropertyType),
		Rooms:        r.Rooms,
		AreaM2:       r.AreaM2,
		Floor:        r.Floor,
		AskingPrice:  r.AskingPrice,
		Currency:     r.Currency,
		PricePerM2:   r.PricePerM2,
		PriceUAH:     r.PriceUAH,
		PriceUSD:     r.PriceUSD,
		Source:       r.Source,
		SourceURL:    r.SourceURL,
		FirstSeenAt:  fmtTime(r.FirstSeenAt),
		PhotoURL:     r.PhotoURL,
		ListingCount: r.ListingCount,
		Elevator:     r.Elevator,
		Parking:      r.Parking,
		Balcony:      r.Balcony,
		SafeRoom:     r.SafeRoom,
	}
	if r.Location != nil {
		c.Location = &geoJSON{Lat: r.Location.Lat, Lng: r.Location.Lng}
	}
	return c
}

func propertyFrom(p domain.Property) propertyJSON {
	out := propertyJSON{
		ID:               p.ID.String(),
		Country:          p.Country,
		City:             p.City,
		District:         p.District,
		Neighborhood:     p.Neighborhood,
		Street:           p.Street,
		HouseNumber:      p.HouseNumber,
		Address:          p.AddressLine(),
		PropertyType:     string(p.PropertyType),
		Rooms:            p.Rooms,
		Bedrooms:         p.Bedrooms,
		Bathrooms:        p.Bathrooms,
		Floor:            p.Floor,
		TotalFloors:      p.TotalFloors,
		AreaM2:           p.AreaM2,
		BalconyAreaM2:    p.BalconyAreaM2,
		ParkingSpaces:    p.ParkingSpaces,
		ConstructionYear: p.ConstructionYear,
		RenovationState:  p.RenovationState,
		Elevator:         p.Elevator,
		Parking:          p.Parking,
		Storage:          p.Storage,
		Balcony:          p.Balcony,
		AirConditioning:  p.AirConditioning,
		SafeRoom:         p.SafeRoom,
		Accessibility:    p.Accessibility,
		FieldProvenance:  p.FieldProvenance,
		CreatedAt:        fmtTime(p.CreatedAt),
		UpdatedAt:        fmtTime(p.UpdatedAt),
	}
	if p.Location != nil {
		out.Location = &geoJSON{Lat: p.Location.Lat, Lng: p.Location.Lng}
	}
	return out
}

func listingFrom(l domain.Listing) listingJSON {
	out := listingJSON{
		ID:              l.ID.String(),
		PropertyID:      l.PropertyID.String(),
		Source:          l.Source,
		SourceListingID: l.SourceListingID,
		SourceURL:       l.SourceURL,
		ListingType:     string(l.ListingType),
		AskingPrice:     l.AskingPrice,
		Currency:        l.Currency,
		PricePerM2:      l.PricePerM2,
		PriceUAH:        l.PriceUAH,
		PriceUSD:        l.PriceUSD,
		PricePerM2UAH:   l.PricePerM2UAH,
		ExchangeRate:    l.ExchangeRate,
		FxSource:        l.FxSource,
		Description:     l.Description,
		SellerName:      l.SellerName,
		SellerType:      string(l.SellerType),
		FirstSeenAt:     fmtTime(l.FirstSeenAt),
		LastSeenAt:      fmtTime(l.LastSeenAt),
		Status:          string(l.Status),
		Photos:          l.Photos,
		MatchConfidence: l.MatchConfidence,
	}
	out.RemovedAt = fmtTimePtr(l.RemovedAt)
	return out
}

func historyFrom(h domain.ListingHistory) historyJSON {
	return historyJSON{
		ID:              h.ID.String(),
		ListingID:       h.ListingID.String(),
		RecordedAt:      fmtTime(h.RecordedAt),
		Price:           h.Price,
		Currency:        h.Currency,
		PriceUAH:        h.PriceUAH,
		PriceUSD:        h.PriceUSD,
		DescriptionHash: h.DescriptionHash,
		Status:          string(h.Status),
		SellerType:      string(h.SellerType),
		PhotosHash:      h.PhotosHash,
	}
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
