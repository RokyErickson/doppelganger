package process

import (
	"testing"
)

func TestExecutableNameWindows(t *testing.T) {
	if name := ExecutableName("doppelganger-agent", "windows"); name != "doppelganger-agent.exe" {
		t.Error("executable name incorrect for Windows")
	}
}

func TestExecutableNameLinux(t *testing.T) {
	if name := ExecutableName("doppelganger-agent", "linux"); name != "doppelganger-agent" {
		t.Error("executable name incorrect for Linux")
	}
}
