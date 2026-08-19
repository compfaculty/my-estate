package stub_test

import (
	"context"
	"testing"

	"github.com/alexg/my-estate/internal/agents/enrichment"
	"github.com/alexg/my-estate/internal/collectors/stub"
	"github.com/alexg/my-estate/internal/domain"
)

func TestStubAgentsEmptyFetch(t *testing.T) {
	ctx := context.Background()
	for _, name := range []string{"lun", "flatfy", "address_ua", "realty100", "facebook", "telegram"} {
		c := stub.New(name)
		if c.Status() != "stub" {
			t.Fatalf("%s status=%s", name, c.Status())
		}
		raws, _, err := c.FetchNewListings(ctx, domain.CrawlCursor{})
		if err != nil {
			t.Fatal(err)
		}
		if len(raws) != 0 {
			t.Fatalf("%s fetch not empty", name)
		}
	}
	if enrichment.Opendatabot().Status() != "stub" || enrichment.YouControl().Status() != "stub" {
		t.Fatal("enrichment stubs")
	}
	if enrichment.Health(ctx) != "stub" {
		t.Fatal("health")
	}
}
