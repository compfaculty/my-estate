package prozorro_test

import (
	"testing"

	"github.com/alexg/my-estate/internal/collectors/prozorro"
)

func TestParseLots(t *testing.T) {
	body := []byte(`{"items":[{"_id":"lot-1","title":"Kyiv lot","status":"active","value":{"amount":100000},"guarantee":{"amount":5000}}]}`)
	lots, err := prozorro.ParseLots(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(lots) != 1 || lots[0].SourceLotID != "lot-1" {
		t.Fatalf("%+v", lots)
	}
	if lots[0].StartPrice == nil || *lots[0].StartPrice != 100000 {
		t.Fatalf("start %v", lots[0].StartPrice)
	}
}
