package enrichment

import (
	"context"

	"github.com/alexg/my-estate/internal/collectors/stub"
)

// Stubs for paid/auth due-diligence. No network calls.
func Opendatabot() *stub.Collector { return stub.New("opendatabot") }
func YouControl() *stub.Collector  { return stub.New("youcontrol") }
func EDR() *stub.Collector         { return stub.New("edr") }
func Cadastre() *stub.Collector    { return stub.New("cadastre") }
func DESSB() *stub.Collector       { return stub.New("edessb") }
func Dream() *stub.Collector       { return stub.New("dream") }

func Health(_ context.Context) string { return "stub" }
