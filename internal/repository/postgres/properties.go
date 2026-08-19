package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type propertyRepo struct{ q querier }

func (r *propertyRepo) Create(ctx context.Context, p *domain.Property) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	lat, lng := locationArgs(p.Location)
	const q = `
		INSERT INTO properties (
			id, country, city, district, neighborhood, street, house_number, location,
			property_type, rooms, bedrooms, bathrooms, floor, total_floors,
			area_m2, balcony_area_m2, lot_area_m2, parking_spaces,
			construction_year, renovation_state, elevator, parking, storage,
			balcony, air_conditioning, safe_room, accessibility, field_provenance,
			oblast, raion, hromada, cadastral_number
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			CASE WHEN $8::float8 IS NULL OR $9::float8 IS NULL THEN NULL
			     ELSE ST_SetSRID(ST_MakePoint($9, $8), 4326)::geography END,
			$10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29,
			$30, $31, $32, $33
		)
		RETURNING created_at, updated_at`
	return r.q.QueryRow(ctx, q,
		p.ID, p.Country, p.City, nullString(p.District), nullString(p.Neighborhood),
		nullString(p.Street), nullString(p.HouseNumber), lat, lng,
		p.PropertyType, p.Rooms, p.Bedrooms, p.Bathrooms, p.Floor, p.TotalFloors,
		p.AreaM2, p.BalconyAreaM2, p.LotAreaM2, p.ParkingSpaces,
		p.ConstructionYear, nullString(p.RenovationState), p.Elevator, p.Parking, p.Storage,
		p.Balcony, p.AirConditioning, p.SafeRoom, p.Accessibility, provenanceJSON(p.FieldProvenance),
		nullString(p.Oblast), nullString(p.Raion), nullString(p.Hromada), nullString(p.CadastralNumber),
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

func (r *propertyRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Property, error) {
	row := r.q.QueryRow(ctx, `SELECT `+propertySelectCols+` FROM properties p WHERE p.id = $1`, id)
	p, err := scanProperty(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *propertyRepo) FindMatch(ctx context.Context, candidate domain.Property) (*domain.Property, error) {
	lat, lng := locationArgs(candidate.Location)
	const q = `
		SELECT ` + propertySelectCols + `
		FROM properties p
		WHERE lower(trim(p.city)) = lower(trim($1))
		  AND lower(coalesce(trim(p.street), '')) = lower(coalesce(trim($2), ''))
		  AND lower(coalesce(trim(p.house_number), '')) = lower(coalesce(trim($3), ''))
		  AND p.property_type = $4
		  AND p.rooms IS NOT DISTINCT FROM $5
		  AND p.area_m2 IS NOT NULL AND $6::numeric IS NOT NULL
		  AND p.area_m2 BETWEEN $6 * 0.95 AND $6 * 1.05
		  AND (
		        p.location IS NULL OR $7::float8 IS NULL OR $8::float8 IS NULL
		        OR ST_DWithin(
		            p.location,
		            ST_SetSRID(ST_MakePoint($8, $7), 4326)::geography,
		            50
		        )
		      )
		LIMIT 1`
	row := r.q.QueryRow(ctx, q,
		candidate.City, candidate.Street, candidate.HouseNumber,
		candidate.PropertyType, candidate.Rooms, candidate.AreaM2, lat, lng,
	)
	p, err := scanProperty(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *propertyRepo) Search(ctx context.Context, filter domain.SearchFilter) ([]domain.PropertySearchResult, int, error) {
	filter = filter.WithDefaults()
	where, args := searchWhere(filter)

	countSQL := `
		SELECT count(*)
		FROM properties p
		JOIN LATERAL (
			SELECT l.*
			FROM listings l
			WHERE l.property_id = p.id
			ORDER BY
				CASE WHEN l.status = 'active' THEN 0 ELSE 1 END,
				l.asking_price ASC NULLS LAST,
				l.last_seen_at DESC
			LIMIT 1
		) cl ON TRUE
		` + where

	var total int
	if err := r.q.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count properties: %w", err)
	}

	limitPH := len(args) + 1
	offsetPH := len(args) + 2
	args = append(args, filter.Limit, filter.Offset)
	listSQL := fmt.Sprintf(`
		SELECT %s,
			cl.asking_price, cl.currency, cl.price_per_m2, cl.source, cl.source_url,
			cl.first_seen_at, cl.photos, cl.id,
			(SELECT count(*) FROM listings lx WHERE lx.property_id = p.id),
			cl.price_uah, cl.price_usd
		FROM properties p
		JOIN LATERAL (
			SELECT l.*
			FROM listings l
			WHERE l.property_id = p.id
			ORDER BY
				CASE WHEN l.status = 'active' THEN 0 ELSE 1 END,
				l.asking_price ASC NULLS LAST,
				l.last_seen_at DESC
			LIMIT 1
		) cl ON TRUE
		%s
		ORDER BY cl.first_seen_at DESC
		LIMIT $%d OFFSET $%d`, propertySelectCols, where, limitPH, offsetPH)

	rows, err := r.q.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search properties: %w", err)
	}
	defer rows.Close()

	out := make([]domain.PropertySearchResult, 0, filter.Limit)
	for rows.Next() {
		var (
			p                              domain.Property
			lat, lng                       *float64
			district, neighborhood, street *string
			house, reno                    *string
			oblast, raion, hromada, cad    *string
			provenance                     []byte
			photos                         []byte
			res                            domain.PropertySearchResult
			listingID                      uuid.UUID
		)
		err := rows.Scan(
			&p.ID, &p.Country, &p.City, &district, &neighborhood, &street, &house,
			&lat, &lng,
			&p.PropertyType, &p.Rooms, &p.Bedrooms, &p.Bathrooms, &p.Floor, &p.TotalFloors,
			&p.AreaM2, &p.BalconyAreaM2, &p.LotAreaM2, &p.ParkingSpaces,
			&p.ConstructionYear, &reno, &p.Elevator, &p.Parking, &p.Storage,
			&p.Balcony, &p.AirConditioning, &p.SafeRoom, &p.Accessibility,
			&oblast, &raion, &hromada, &cad,
			&provenance, &p.CreatedAt, &p.UpdatedAt,
			&res.AskingPrice, &res.Currency, &res.PricePerM2, &res.Source, &res.SourceURL,
			&res.FirstSeenAt, &photos, &listingID, &res.ListingCount,
			&res.PriceUAH, &res.PriceUSD,
		)
		if err != nil {
			return nil, 0, err
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
		var photoList []domain.Photo
		if len(photos) > 0 {
			_ = json.Unmarshal(photos, &photoList)
		}
		if len(photoList) > 0 {
			res.PhotoURL = photoList[0].URL
		}
		res.Property = p
		res.PrimaryListingID = listingID.String()
		out = append(out, res)
	}
	return out, total, rows.Err()
}

func searchWhere(filter domain.SearchFilter) (string, []any) {
	clauses := make([]string, 0, 16)
	args := make([]any, 0, 16)
	push := func(fragment string, vals ...any) {
		for range vals {
			args = append(args, nil)
		}
		start := len(args) - len(vals) + 1
		for i, v := range vals {
			args[start-1+i] = v
		}
		formatted := fragment
		for i := range vals {
			placeholder := fmt.Sprintf("$%d", start+i)
			formatted = strings.Replace(formatted, "?", placeholder, 1)
		}
		clauses = append(clauses, formatted)
	}

	if filter.City != "" {
		push("p.city ILIKE '%' || ? || '%'", filter.City)
	}
	if filter.Neighborhood != "" {
		push("p.neighborhood ILIKE '%' || ? || '%'", filter.Neighborhood)
	}
	if filter.Street != "" {
		push("p.street ILIKE '%' || ? || '%'", filter.Street)
	}
	if filter.Query != "" {
		push(`(
			p.city ILIKE '%' || ? || '%' OR
			p.neighborhood ILIKE '%' || ? || '%' OR
			p.street ILIKE '%' || ? || '%' OR
			cl.description ILIKE '%' || ? || '%'
		)`, filter.Query, filter.Query, filter.Query, filter.Query)
	}
	if filter.MinPrice != nil {
		push("COALESCE(cl.price_uah, cl.asking_price) >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		push("COALESCE(cl.price_uah, cl.asking_price) <= ?", *filter.MaxPrice)
	}
	if filter.MinArea != nil {
		push("p.area_m2 >= ?", *filter.MinArea)
	}
	if filter.MaxArea != nil {
		push("p.area_m2 <= ?", *filter.MaxArea)
	}
	if filter.MinRooms != nil {
		push("p.rooms >= ?", *filter.MinRooms)
	}
	if filter.MaxRooms != nil {
		push("p.rooms <= ?", *filter.MaxRooms)
	}
	if filter.MinFloor != nil {
		push("p.floor >= ?", *filter.MinFloor)
	}
	if filter.MaxFloor != nil {
		push("p.floor <= ?", *filter.MaxFloor)
	}
	if filter.PropertyType != "" {
		push("p.property_type = ?", string(filter.PropertyType))
	}
	if filter.Source != "" {
		push("cl.source = ?", filter.Source)
	}
	if filter.ListingAge != nil {
		push("cl.first_seen_at >= now() - ?::interval", fmt.Sprintf("%d seconds", int(filter.ListingAge.Seconds())))
	}
	if filter.Elevator != nil {
		if *filter.Elevator {
			clauses = append(clauses, "p.elevator IS TRUE")
		} else {
			clauses = append(clauses, "p.elevator IS NOT TRUE")
		}
	}
	if filter.Parking != nil {
		if *filter.Parking {
			clauses = append(clauses, "p.parking IS TRUE")
		} else {
			clauses = append(clauses, "p.parking IS NOT TRUE")
		}
	}
	if filter.Balcony != nil {
		if *filter.Balcony {
			clauses = append(clauses, "p.balcony IS TRUE")
		} else {
			clauses = append(clauses, "p.balcony IS NOT TRUE")
		}
	}
	if filter.SafeRoom != nil {
		if *filter.SafeRoom {
			clauses = append(clauses, "p.safe_room IS TRUE")
		} else {
			clauses = append(clauses, "p.safe_room IS NOT TRUE")
		}
	}
	if filter.MinLat != nil && filter.MinLng != nil && filter.MaxLat != nil && filter.MaxLng != nil {
		push(
			"p.location IS NOT NULL AND ST_Intersects(p.location::geometry, ST_MakeEnvelope(?, ?, ?, ?, 4326))",
			*filter.MinLng, *filter.MinLat, *filter.MaxLng, *filter.MaxLat,
		)
	}
	if filter.Lat != nil && filter.Lng != nil && filter.RadiusM != nil {
		push(
			"p.location IS NOT NULL AND ST_DWithin(p.location, ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
			*filter.Lng, *filter.Lat, *filter.RadiusM,
		)
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
