package dimria_test

import (
	"testing"

	"github.com/alexg/my-estate/internal/collectors/dimria"
)

func TestParseSearchIDs(t *testing.T) {
	ids, err := dimria.ParseSearchIDs([]byte(`{"items":[101,102,"103"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 3 || ids[0] != "101" || ids[2] != "103" {
		t.Fatalf("%v", ids)
	}
}
