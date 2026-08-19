package dedup

import "testing"

func TestNormalizeAddress(t *testing.T) {
	cases := map[string]string{
		"  Herzl  ":        "herzl",
		"Bnei-Dan":         "bnei-dan",
		"Rothschild Blvd.": "rothschild blvd",
		"Tel   Aviv":       "tel aviv",
		"  Хорива  ":       "хорива",
		"Оболонський проспект": "оболонський проспект",
	}
	for in, want := range cases {
		if got := NormalizeAddress(in); got != want {
			t.Errorf("NormalizeAddress(%q)=%q want %q", in, got, want)
		}
	}
}

func TestAreaWithinPercent(t *testing.T) {
	if !AreaWithinPercent(75, 76, 5) {
		t.Fatal("75 vs 76 should match within 5%")
	}
	if AreaWithinPercent(75, 90, 5) {
		t.Fatal("75 vs 90 should not match within 5%")
	}
	if AreaWithinPercent(0, 75, 5) {
		t.Fatal("zero area should not match")
	}
}
