package normalize_test

import (
	"testing"

	"github.com/alexg/my-estate/internal/normalize"
)

func TestParseDimriaFixture(t *testing.T) {
	body := []byte(`{
		"realty_id": 555,
		"price": 3200000,
		"currency_type": "UAH",
		"rooms_count": 2,
		"total_square_meters": 64,
		"street_name": "Хрещатик",
		"building_number_str": "10",
		"floor": 4,
		"floors_count": 9,
		"latitude": 50.45,
		"longitude": 30.52,
		"description": "тест"
	}`)
	got, err := normalize.Parse("dimria", body)
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceListingID != "555" {
		t.Fatalf("id=%s", got.SourceListingID)
	}
	if got.Property.City != "Київ" || got.Property.Street != "Хрещатик" {
		t.Fatalf("geo %+v", got.Property)
	}
	if got.AskingPrice == nil || *got.AskingPrice != 3200000 {
		t.Fatalf("price %v", got.AskingPrice)
	}
	if got.Currency != "UAH" {
		t.Fatalf("ccy %s", got.Currency)
	}
}
