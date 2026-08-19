package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PropertyType string

const (
	PropertyTypeApartment  PropertyType = "apartment"
	PropertyTypeHouse      PropertyType = "house"
	PropertyTypeVilla      PropertyType = "villa"
	PropertyTypePenthouse  PropertyType = "penthouse"
	PropertyTypeDuplex     PropertyType = "duplex"
	PropertyTypeLand       PropertyType = "land"
	PropertyTypeCommercial PropertyType = "commercial"
)

type ListingType string

const (
	ListingTypeSale ListingType = "sale"
	ListingTypeRent ListingType = "rent"
)

type SellerType string

const (
	SellerTypePrivate SellerType = "private"
	SellerTypeBroker  SellerType = "broker"
	SellerTypeUnknown SellerType = "unknown"
)

type ListingStatus string

const (
	ListingStatusActive  ListingStatus = "active"
	ListingStatusRemoved ListingStatus = "removed"
	ListingStatusUnknown ListingStatus = "unknown"
)

type FieldProvenance struct {
	Field      string  `json:"field"`
	Source     string  `json:"source"`
	Confidence float64 `json:"confidence"`
}

type Photo struct {
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

type GeoPoint struct {
	Lat float64
	Lng float64
}

type Property struct {
	ID               uuid.UUID
	Country          string
	Oblast           string
	Raion            string
	Hromada          string
	City             string
	District         string
	Neighborhood     string
	Street           string
	HouseNumber      string
	CadastralNumber  string
	Location         *GeoPoint
	PropertyType     PropertyType
	Rooms            *float64
	Bedrooms         *int
	Bathrooms        *float64
	Floor            *int
	TotalFloors      *int
	AreaM2           *float64
	BalconyAreaM2    *float64
	LotAreaM2        *float64
	ParkingSpaces    *int
	ConstructionYear *int
	RenovationState  string
	Elevator         *bool
	Parking          *bool
	Storage          *bool
	Balcony          *bool
	AirConditioning  *bool
	SafeRoom         *bool
	Accessibility    *bool
	FieldProvenance  []FieldProvenance
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (p Property) AddressLine() string {
	parts := make([]string, 0, 4)
	if p.Street != "" {
		line := p.Street
		if p.HouseNumber != "" {
			line += " " + p.HouseNumber
		}
		parts = append(parts, line)
	}
	if p.Neighborhood != "" {
		parts = append(parts, p.Neighborhood)
	}
	if p.City != "" {
		parts = append(parts, p.City)
	}
	out := ""
	for i, part := range parts {
		if i > 0 {
			out += ", "
		}
		out += part
	}
	return out
}

type Listing struct {
	ID               uuid.UUID
	PropertyID       uuid.UUID
	Source           string
	SourceListingID  string
	SourceURL        string
	ListingType      ListingType
	AskingPrice      *float64
	Currency         string
	PricePerM2       *float64
	PriceOriginal    *float64
	CurrencyOriginal string
	ExchangeRate     *float64
	ExchangeRateDate *time.Time
	FxSource         string
	PriceUAH         *float64
	PriceUSD         *float64
	PricePerM2UAH    *float64
	Description      string
	DescriptionHash  string
	SellerName       string
	SellerType       SellerType
	SellerPhoneHash  string
	FirstSeenAt      time.Time
	LastSeenAt       time.Time
	RemovedAt        *time.Time
	Status           ListingStatus
	Photos           []Photo
	PhotosHash       string
	Metadata         json.RawMessage
	MatchConfidence  *float64
}

func (l Listing) Snapshot() ListingHistory {
	return ListingHistory{
		ListingID:        l.ID,
		Price:            l.AskingPrice,
		Currency:         l.Currency,
		PriceOriginal:    l.PriceOriginal,
		CurrencyOriginal: l.CurrencyOriginal,
		PriceUAH:         l.PriceUAH,
		PriceUSD:         l.PriceUSD,
		ExchangeRate:     l.ExchangeRate,
		ExchangeRateDate: l.ExchangeRateDate,
		DescriptionHash:  l.DescriptionHash,
		Status:           l.Status,
		SellerType:       l.SellerType,
		PhotosHash:       l.PhotosHash,
	}
}

func (l Listing) ChangedFrom(prev Listing) bool {
	if !floatPtrEqual(l.AskingPrice, prev.AskingPrice) {
		return true
	}
	if !floatPtrEqual(l.PriceOriginal, prev.PriceOriginal) {
		return true
	}
	if !floatPtrEqual(l.PriceUAH, prev.PriceUAH) {
		return true
	}
	if l.CurrencyOriginal != prev.CurrencyOriginal {
		return true
	}
	if l.DescriptionHash != prev.DescriptionHash {
		return true
	}
	if l.Status != prev.Status {
		return true
	}
	if l.SellerType != prev.SellerType {
		return true
	}
	if l.PhotosHash != prev.PhotosHash {
		return true
	}
	return false
}

type ListingHistory struct {
	ID               uuid.UUID
	ListingID        uuid.UUID
	RecordedAt       time.Time
	Price            *float64
	Currency         string
	PriceOriginal    *float64
	CurrencyOriginal string
	PriceUAH         *float64
	PriceUSD         *float64
	ExchangeRate     *float64
	ExchangeRateDate *time.Time
	DescriptionHash  string
	Status           ListingStatus
	SellerType       SellerType
	PhotosHash       string
}

type FxRate struct {
	RateDate  time.Time
	CCY       string
	RateUAH   float64
	FetchedAt time.Time
	Source    string
}

type MarketStatistic struct {
	ID         uuid.UUID
	Source     string
	Region     string
	Market     string
	Period     string
	IndexValue *float64
	Payload    json.RawMessage
	FetchedAt  time.Time
}

type AuctionLot struct {
	ID               uuid.UUID
	Source           string
	SourceLotID      string
	SourceURL        string
	Title            string
	Description      string
	City             string
	Oblast           string
	StartPrice       *float64
	Currency         string
	GuaranteeDeposit *float64
	AuctionStage     string
	DeadlineAt       *time.Time
	Organizer        string
	Protocol         string
	FinalPrice       *float64
	RawPayload       json.RawMessage
	FirstSeenAt      time.Time
	LastSeenAt       time.Time
}

type RawListing struct {
	ID               uuid.UUID
	Source           string
	SourceListingID  string
	SourceURL        string
	FetchedAt        time.Time
	RawPayload       json.RawMessage
	ContentHash      string
	HTTPStatus       *int
	CollectorVersion string
}

type Source struct {
	ID                   uuid.UUID
	Name                 string
	Enabled              bool
	CrawlIntervalSeconds int
	Concurrency          int
	AuthMode             string
	FetchMode            string
	LastSuccessAt        *time.Time
	LastError            *string
	LastCrawlAt          *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ListingsDiscovered   int
	ListingsUpdated      int
	LastRunStatus        string
}

type CollectorRun struct {
	ID                 uuid.UUID
	SourceID           uuid.UUID
	StartedAt          time.Time
	FinishedAt         *time.Time
	Status             string
	ListingsDiscovered int
	ListingsUpdated    int
	ListingsRemoved    int
	Error              *string
	Cursor             json.RawMessage
}

type CrawlCursor struct {
	Page int `json:"page"`
}

type EventType string

const (
	EventListingDiscovered EventType = "listing.discovered"
	EventListingUpdated    EventType = "listing.updated"
	EventListingRemoved    EventType = "listing.removed"
)

type Event struct {
	Type       EventType
	ListingID  uuid.UUID
	PropertyID uuid.UUID
	At         time.Time
}

func floatPtrEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
