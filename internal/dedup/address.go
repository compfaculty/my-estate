package dedup

import (
	"regexp"
	"strings"
	"unicode"
)

var spaceRE = regexp.MustCompile(`\s+`)

// NormalizeAddress folds whitespace and case for deterministic matching.
func NormalizeAddress(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) && r != '-' {
			return ' '
		}
		return r
	}, s)
	return strings.TrimSpace(spaceRE.ReplaceAllString(s, " "))
}

func AreaWithinPercent(a, b, pct float64) bool {
	if a <= 0 || b <= 0 {
		return false
	}
	lo, hi := a, b
	if lo > hi {
		lo, hi = hi, lo
	}
	return (hi-lo)/lo <= pct/100
}
