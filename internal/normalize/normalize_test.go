package normalize_test

import (
	"encoding/json"
	"testing"

	"github.com/alexg/my-estate/internal/collectors/mock"
	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/hashutil"
	"github.com/alexg/my-estate/internal/normalize"
)

func TestParseMockListing(t *testing.T) {
	payload := mock.Payload{
		SourceListingID: "x-1",
		SourceURL:       "https://mock.local/listings/x-1",
		City:            "Київ",
		Neighborhood:    "Поділ",
		Street:          "Хорива",
		HouseNumber:     "12",
		Lat:             50.4691,
		Lng:             30.5134,
		PropertyType:    "apartment",
		Rooms:           3,
		Bedrooms:        2,
		Bathrooms:       1,
		Floor:           2,
		TotalFloors:     5,
		AreaM2:          75,
		ListingType:     "sale",
		Price:           2680000,
		Currency:        "UAH",
		Description:     "Test apartment",
		SellerName:      "Dana",
		SellerType:      "private",
		Phone:           "050-111-1001",
		Photos:          []string{"https://example.com/a.jpg"},
		Status:          "active",
		Elevator:        true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	got, err := normalize.Parse("mock", body)
	if err != nil {
		t.Fatal(err)
	}
	if got.Property.City != "Київ" {
		t.Fatalf("city=%q", got.Property.City)
	}
	if got.Property.PropertyType != domain.PropertyTypeApartment {
		t.Fatalf("type=%s", got.Property.PropertyType)
	}
	if got.AskingPrice == nil || *got.AskingPrice != 2680000 {
		t.Fatalf("price=%v", got.AskingPrice)
	}
	if got.SellerPhoneHash != hashutil.PhoneHash("0501111001") {
		t.Fatalf("phone hash mismatch: %s", got.SellerPhoneHash)
	}
	ppm2 := got.PricePerM2()
	if ppm2 == nil || *ppm2 < 35730 || *ppm2 > 35734 {
		t.Fatalf("price/m2=%v", ppm2)
	}
	if got.DescriptionHash() == "" || got.PhotosHash() == "" {
		t.Fatal("expected hashes")
	}
}

func TestParseUnsupportedSource(t *testing.T) {
	_, err := normalize.Parse("yad2", []byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
