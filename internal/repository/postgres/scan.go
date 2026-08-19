package postgres

import (
	"encoding/json"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/jackc/pgx/v5"
)

const propertySelectCols = `
	p.id, p.country, p.city, p.district, p.neighborhood, p.street, p.house_number,
	ST_Y(p.location::geometry), ST_X(p.location::geometry),
	p.property_type, p.rooms, p.bedrooms, p.bathrooms, p.floor, p.total_floors,
	p.area_m2, p.balcony_area_m2, p.lot_area_m2, p.parking_spaces,
	p.construction_year, p.renovation_state, p.elevator, p.parking, p.storage,
	p.balcony, p.air_conditioning, p.safe_room, p.accessibility,
	p.oblast, p.raion, p.hromada, p.cadastral_number,
	p.field_provenance, p.created_at, p.updated_at`

func scanProperty(row pgx.Row) (*domain.Property, error) {
	var (
		p                              domain.Property
		lat, lng                       *float64
		district, neighborhood, street *string
		house, reno                    *string
		oblast, raion, hromada, cad    *string
		provenance                     []byte
	)
	err := row.Scan(
		&p.ID, &p.Country, &p.City, &district, &neighborhood, &street, &house,
		&lat, &lng,
		&p.PropertyType, &p.Rooms, &p.Bedrooms, &p.Bathrooms, &p.Floor, &p.TotalFloors,
		&p.AreaM2, &p.BalconyAreaM2, &p.LotAreaM2, &p.ParkingSpaces,
		&p.ConstructionYear, &reno, &p.Elevator, &p.Parking, &p.Storage,
		&p.Balcony, &p.AirConditioning, &p.SafeRoom, &p.Accessibility,
		&oblast, &raion, &hromada, &cad,
		&provenance, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.District = deref(district)
	p.Neighborhood = deref(neighborhood)
	p.Street = deref(street)
	p.HouseNumber = deref(house)
	p.RenovationState = deref(reno)
	p.Oblast = deref(oblast)
	p.Raion = deref(raion)
	p.Hromada = deref(hromada)
	p.CadastralNumber = deref(cad)
	if lat != nil && lng != nil {
		p.Location = &domain.GeoPoint{Lat: *lat, Lng: *lng}
	}
	if len(provenance) > 0 {
		_ = json.Unmarshal(provenance, &p.FieldProvenance)
	}
	return &p, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func locationArgs(loc *domain.GeoPoint) (lat any, lng any) {
	if loc == nil {
		return nil, nil
	}
	return loc.Lat, loc.Lng
}

func provenanceJSON(p []domain.FieldProvenance) []byte {
	if p == nil {
		p = []domain.FieldProvenance{}
	}
	b, err := json.Marshal(p)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func photosJSON(photos []domain.Photo) []byte {
	if photos == nil {
		photos = []domain.Photo{}
	}
	b, err := json.Marshal(photos)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func metadataJSON(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}
