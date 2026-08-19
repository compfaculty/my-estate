package olx_test

import (
	"testing"

	"github.com/alexg/my-estate/internal/collectors/olx"
)

func TestParseExtractFixtureHTMLJSON(t *testing.T) {
	html := `<html><body><h1>Квартира на Подолі</h1><div>2 100 000 грн</div></body></html>`
	p := olx.ParseExtract("https://www.olx.ua/d/uk/obyavlenie/kvartyra-IDABC.html", olx.Extract{
		HTML: html,
		JSON: map[string]any{
			"id":       "IDABC",
			"title":    "Квартира на Подолі",
			"price":    2100000.0,
			"currency": "UAH",
			"hrefs":    []any{"https://www.olx.ua/d/uk/obyavlenie/other-ID999.html"},
		},
	})
	if p.Price != 2100000 {
		t.Fatalf("price=%v", p.Price)
	}
	if p.Currency != "UAH" {
		t.Fatalf("ccy=%s", p.Currency)
	}
	if p.SourceListingID != "IDABC" {
		t.Fatalf("id=%s", p.SourceListingID)
	}
	if p.City != "Київ" {
		t.Fatalf("city=%s", p.City)
	}
}

func TestAllowlistRejects(t *testing.T) {
	c := olx.New()
	_, err := c.FetchListingDetails(t.Context(), "https://evil.example/x")
	if err == nil {
		t.Fatal("expected allowlist error")
	}
}
