package rieltor_test

import (
	"testing"

	"github.com/alexg/my-estate/internal/collectors/rieltor"
)

func TestParseListingHTML(t *testing.T) {
	html := []byte(`<html><script type="application/ld+json">{"name":"Квартира"}</script><div>"price": "1550000"</div></html>`)
	p := rieltor.ParseListingHTML("https://rieltor.ua/kiev/flats-sale/999/", html)
	if p.SourceListingID != "999" {
		t.Fatalf("id=%s", p.SourceListingID)
	}
	if p.Price != 1550000 {
		t.Fatalf("price=%v", p.Price)
	}
	if p.City != "Київ" {
		t.Fatalf("city=%s", p.City)
	}
}
