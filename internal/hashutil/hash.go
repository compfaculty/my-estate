package hashutil

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"
)

func SHA256Hex(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func PhoneHash(phone string) string {
	var b strings.Builder
	for _, r := range phone {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return ""
	}
	return SHA256Hex(b.String())
}
