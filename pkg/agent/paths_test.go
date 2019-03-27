package agent

import (
	"strings"
	"testing"

	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
)

func TestInstallPath(t *testing.T) {

	if p, err := installPath(); err != nil {
		t.Fatal("unable to compute/create install path:", err)
	} else if p == "" {
		t.Error("empty install path returned")
	} else if !strings.Contains(p, doppelganger.Version) {
		t.Error("install path does not contain Doppelganger version")
	}
}
