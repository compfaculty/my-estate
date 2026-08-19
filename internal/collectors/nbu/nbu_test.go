package nbu_test

import (
	"testing"
	"time"

	"github.com/alexg/my-estate/internal/collectors/nbu"
)

func TestParseRatesUSDAndEUR(t *testing.T) {
	body := []byte(`[
		{"cc":"USD","rate":41.2,"exchangedate":"15.01.2024"},
		{"cc":"EUR","rate":45.1,"exchangedate":"15.01.2024"},
		{"cc":"PLN","rate":10.1,"exchangedate":"15.01.2024"}
	]`)
	rates, err := nbu.ParseRates(body, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(rates) != 2 {
		t.Fatalf("got %d rates", len(rates))
	}
}

func TestConvertUSDDates(t *testing.T) {
	uah1, usd1 := nbu.Convert(100000, "USD", 40)
	uah2, usd2 := nbu.Convert(100000, "USD", 42)
	if usd1 != 100000 || usd2 != 100000 {
		t.Fatal("USD original must stay")
	}
	if uah1 == uah2 {
		t.Fatal("UAH must differ across rates")
	}
	if uah1 != 4_000_000 || uah2 != 4_200_000 {
		t.Fatalf("got %v %v", uah1, uah2)
	}
}
