package mock

// Payload is the mock source JSON shape stored in raw_listings.raw_payload.
type Payload struct {
	SourceListingID  string   `json:"source_listing_id"`
	SourceURL        string   `json:"source_url"`
	City             string   `json:"city"`
	District         string   `json:"district,omitempty"`
	Neighborhood     string   `json:"neighborhood"`
	Street           string   `json:"street"`
	HouseNumber      string   `json:"house_number"`
	Lat              float64  `json:"lat"`
	Lng              float64  `json:"lng"`
	PropertyType     string   `json:"property_type"`
	Rooms            float64  `json:"rooms"`
	Bedrooms         int      `json:"bedrooms"`
	Bathrooms        float64  `json:"bathrooms"`
	Floor            int      `json:"floor"`
	TotalFloors      int      `json:"total_floors"`
	AreaM2           float64  `json:"area_m2"`
	BalconyAreaM2    float64  `json:"balcony_area_m2,omitempty"`
	ParkingSpaces    int      `json:"parking_spaces,omitempty"`
	ConstructionYear int      `json:"construction_year,omitempty"`
	RenovationState  string   `json:"renovation_state,omitempty"`
	Elevator         bool     `json:"elevator"`
	Parking          bool     `json:"parking"`
	Storage          bool     `json:"storage"`
	Balcony          bool     `json:"balcony"`
	AirConditioning  bool     `json:"air_conditioning"`
	SafeRoom         bool     `json:"safe_room"`
	Accessibility    bool     `json:"accessibility"`
	ListingType      string   `json:"listing_type"`
	Price            float64  `json:"price"`
	Currency         string   `json:"currency"`
	Description      string   `json:"description"`
	SellerName       string   `json:"seller_name"`
	SellerType       string   `json:"seller_type"`
	Phone            string   `json:"phone"`
	Photos           []string `json:"photos"`
	Status           string   `json:"status"`
}
