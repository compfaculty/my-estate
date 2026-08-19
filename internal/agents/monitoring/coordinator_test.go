package monitoring_test

import (
	"testing"

	"github.com/alexg/my-estate/internal/agents/monitoring"
	"github.com/alexg/my-estate/internal/config"
)

func TestCoordinatorStubsWithoutLiveKeys(t *testing.T) {
	c := monitoring.NewCoordinator(config.Config{}, nil, nil, nil)
	names := map[string]bool{}
	for _, a := range c.Agents {
		names[a.Name()] = true
	}
	for _, n := range []string{"lun", "flatfy", "facebook", "telegram", "youcontrol", "opendatabot", "nbu", "derzhstat", "prozorro", "rieltor"} {
		if !names[n] {
			t.Fatalf("missing agent %s", n)
		}
	}
	if names["olx"] || names["dimria"] {
		t.Fatal("olx/dimria should stay off without flags/keys")
	}
}
