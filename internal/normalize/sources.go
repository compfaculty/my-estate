package normalize

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/alexg/my-estate/internal/collectors/dimria"
	"github.com/alexg/my-estate/internal/collectors/mock"
	"github.com/alexg/my-estate/internal/collectors/olx"
	"github.com/alexg/my-estate/internal/collectors/rieltor"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/hashutil"
)

// Parse converts a source-specific raw payload into a canonical listing.
func Parse(source string, payload json.RawMessage) (Listing, error) {
	switch source {
	case mock.Name:
		return parseMock(payload)
	case dimria.Name:
		return parseDimria(payload)
	case rieltor.Name:
		return parseRieltor(payload)
	case olx.Name:
		return parseOLX(payload)
	default:
		return Listing{}, fmt.Errorf("unsupported source %q", source)
	}
}

func kyivBase(source string) domain.Property {
	return domain.Property{
		Country:      "UA",
		Oblast:       "Київ",
		City:         "Київ",
		PropertyType: domain.PropertyTypeApartment,
		FieldProvenance: []domain.FieldProvenance{
			{Field: "city", Source: source, Confidence: 1},
		},
	}
}

func parseDimria(payload json.RawMessage) (Listing, error) {
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return Listing{}, err
	}
	id := fmt.Sprint(first(m, "realty_id", "realtyId", "id"))
	price, _ := asF(first(m, "price", "price_total"))
	ccy := dimriaCurrency(first(m, "currency_type", "currency_type_id", "currency"))
	rooms, _ := asF(first(m, "rooms_count", "rooms"))
	area, _ := asF(first(m, "total_square_meters", "square_total"))
	floor, _ := asI(first(m, "floor"))
	total, _ := asI(first(m, "floors_count", "floors"))
	lat, _ := asF(first(m, "latitude", "lat"))
	lng, _ := asF(first(m, "longitude", "lng"))
	prop := kyivBase(dimria.Name)
	prop.District = str(first(m, "district_name", "district"))
	prop.Street = str(first(m, "street_name", "street"))
	prop.HouseNumber = str(first(m, "building_number_str", "building_number"))
	prop.Rooms = &rooms
	prop.AreaM2 = &area
	prop.Floor = &floor
	prop.TotalFloors = &total
	if lat != 0 && lng != 0 {
		prop.Location = &domain.GeoPoint{Lat: lat, Lng: lng}
	}
	beautiful := str(first(m, "beautiful_url", "url"))
	srcURL := "https://dom.ria.com/uk/" + id
	if beautiful != "" {
		if strings.HasPrefix(beautiful, "http") {
			srcURL = beautiful
		} else {
			srcURL = "https://dom.ria.com" + beautiful
		}
	}
	return Listing{
		Property:        prop,
		Source:          dimria.Name,
		SourceListingID: id,
		SourceURL:       srcURL,
		ListingType:     domain.ListingTypeSale,
		AskingPrice:     &price,
		Currency:        ccy,
		Description:     str(first(m, "description", "title")),
		Status:          domain.ListingStatusActive,
		Metadata:        payload,
	}, nil
}

func parseRieltor(payload json.RawMessage) (Listing, error) {
	var p rieltor.Payload
	if err := json.Unmarshal(payload, &p); err != nil {
		return Listing{}, err
	}
	prop := kyivBase(rieltor.Name)
	prop.District = p.District
	prop.Street = p.Street
	prop.HouseNumber = p.HouseNumber
	prop.Rooms = &p.Rooms
	prop.AreaM2 = &p.AreaM2
	prop.Floor = &p.Floor
	prop.TotalFloors = &p.TotalFloors
	price := p.Price
	return Listing{
		Property:        prop,
		Source:          rieltor.Name,
		SourceListingID: p.SourceListingID,
		SourceURL:       p.SourceURL,
		ListingType:     domain.ListingTypeSale,
		AskingPrice:     &price,
		Currency:        nz(p.Currency, "UAH"),
		Description:     p.Description,
		Status:          domain.ListingStatusActive,
		Metadata:        payload,
	}, nil
}

func parseOLX(payload json.RawMessage) (Listing, error) {
	var p olx.Payload
	if err := json.Unmarshal(payload, &p); err != nil {
		return Listing{}, err
	}
	prop := kyivBase(olx.Name)
	prop.Street = p.Street
	prop.HouseNumber = p.HouseNumber
	prop.Rooms = &p.Rooms
	prop.AreaM2 = &p.AreaM2
	prop.Floor = &p.Floor
	price := p.Price
	return Listing{
		Property:        prop,
		Source:          olx.Name,
		SourceListingID: p.SourceListingID,
		SourceURL:       p.SourceURL,
		ListingType:     domain.ListingTypeSale,
		AskingPrice:     &price,
		Currency:        nz(p.Currency, "UAH"),
		Description:     p.Description,
		SellerPhoneHash: hashutil.PhoneHash(""),
		Status:          domain.ListingStatusActive,
		Metadata:        payload,
	}, nil
}

func dimriaCurrency(v any) string {
	switch t := v.(type) {
	case string:
		s := strings.ToUpper(t)
		if strings.Contains(s, "USD") || s == "1" {
			return "USD"
		}
		return "UAH"
	case float64:
		if int(t) == 1 {
			return "USD"
		}
		return "UAH"
	default:
		return "UAH"
	}
}

func first(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func str(v any) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func asF(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case string:
		f, err := strconv.ParseFloat(strings.ReplaceAll(t, " ", ""), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func asI(v any) (int, bool) {
	f, ok := asF(v)
	return int(f), ok
}

func nz(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
