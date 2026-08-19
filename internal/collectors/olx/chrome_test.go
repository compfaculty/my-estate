package olx

import "testing"

func TestHTMLBlocked(t *testing.T) {
	if !htmlBlocked(`<html>please solve captcha</html>`) {
		t.Fatal("expected captcha block")
	}
	if htmlBlocked(`<html><h1>Квартира</h1></html>`) {
		t.Fatal("clean html")
	}
}

func TestFindChromeDoesNotPanic(t *testing.T) {
	_ = FindChrome()
}
