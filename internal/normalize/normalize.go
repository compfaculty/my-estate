package normalize

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alexg/my-estate/internal/collectors/mock"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/hashutil"
)

// Listing is the canonical ingest input produced from a raw source payload.
type Listing struct {
	Property         domain.Property
	Source           string
	SourceListingID  string
	SourceURL        string
	ListingType      domain.ListingType
	AskingPrice      *float64
	Currency         string
	PriceOriginal    *float64
	CurrencyOriginal string
	ExchangeRate     *float64
	ExchangeRateDate *time.Time
	FxSource         string
	PriceUAH         *float64
	PriceUSD         *float64
	PricePerM2UAH    *float64
	Description      string
	SellerName       string
	SellerType       domain.SellerType
	SellerPhoneHash  string
	Status           domain.ListingStatus
	Photos           []domain.Photo
	Metadata         json.RawMessage
}

func (n Listing) DescriptionHash() string {
	return hashutil.SHA256Hex(n.Description)
}

func (n Listing) PhotosHash() string {
	parts := make([]string, 0, len(n.Photos))
	for _, p := range n.Photos {
		parts = append(parts, p.URL)
	}
	return hashutil.SHA256Hex(parts...)
}

func (n Listing) PricePerM2() *float64 {
	if n.AskingPrice == nil || n.Property.AreaM2 == nil || *n.Property.AreaM2 == 0 {
		return nil
	}
	v := *n.AskingPrice / *n.Property.AreaM2
	return &v
}

func parseMock(payload json.RawMessage) (Listing, error) {
	var p mock.Payload
	if err := json.Unmarshal(payload, &p); err != nil {
		return Listing{}, fmt.Errorf("mock payload: %w", err)
	}
	rooms := p.Rooms
	area := p.AreaM2
	bed := p.Bedrooms
	bath := p.Bathrooms
	floor := p.Floor
	total := p.TotalFloors
	price := p.Price
	year := p.ConstructionYear
	balconyArea := p.BalconyAreaM2
	parkingSpaces := p.ParkingSpaces
	elev := p.Elevator
	park := p.Parking
	stor := p.Storage
	balc := p.Balcony
	ac := p.AirConditioning
	safe := p.SafeRoom
	acc := p.Accessibility

	prop := domain.Property{
		Country:          "UA",
		Oblast:           "Київ",
		City:             strings.TrimSpace(p.City),
		District:         strings.TrimSpace(p.District),
		Neighborhood:     strings.TrimSpace(p.Neighborhood),
		Street:           strings.TrimSpace(p.Street),
		HouseNumber:      strings.TrimSpace(p.HouseNumber),
		Location:         &domain.GeoPoint{Lat: p.Lat, Lng: p.Lng},
		PropertyType:     domain.PropertyType(p.PropertyType),
		Rooms:            &rooms,
		Bedrooms:         &bed,
		Bathrooms:        &bath,
		Floor:            &floor,
		TotalFloors:      &total,
		AreaM2:           &area,
		ConstructionYear: ptrIfNonZero(year),
		RenovationState:  p.RenovationState,
		Elevator:         &elev,
		Parking:          &park,
		Storage:          &stor,
		Balcony:          &balc,
		AirConditioning:  &ac,
		SafeRoom:         &safe,
		Accessibility:    &acc,
		FieldProvenance: []domain.FieldProvenance{
			{Field: "area_m2", Source: mock.Name, Confidence: 1},
			{Field: "rooms", Source: mock.Name, Confidence: 1},
			{Field: "asking_price", Source: mock.Name, Confidence: 1},
		},
	}
	if balconyArea > 0 {
		prop.BalconyAreaM2 = &balconyArea
	}
	if parkingSpaces > 0 {
		prop.ParkingSpaces = &parkingSpaces
	}

	photos := make([]domain.Photo, 0, len(p.Photos))
	for i, url := range p.Photos {
		photos = append(photos, domain.Photo{URL: url, SortOrder: i})
	}

	status := domain.ListingStatus(p.Status)
	if status == "" {
		status = domain.ListingStatusActive
	}
	seller := domain.SellerType(p.SellerType)
	if seller == "" {
		seller = domain.SellerTypeUnknown
	}
	lt := domain.ListingType(p.ListingType)
	if lt == "" {
		lt = domain.ListingTypeSale
	}

	return Listing{
		Property:        prop,
		Source:          mock.Name,
		SourceListingID: p.SourceListingID,
		SourceURL:       p.SourceURL,
		ListingType:     lt,
		AskingPrice:     &price,
		Currency:        p.Currency,
		Description:     p.Description,
		SellerName:      p.SellerName,
		SellerType:      seller,
		SellerPhoneHash: hashutil.PhoneHash(p.Phone),
		Status:          status,
		Photos:          photos,
		Metadata:        payload,
	}, nil
}

func ptrIfNonZero(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}
