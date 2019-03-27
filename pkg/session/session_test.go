package session

import (
	"testing"

	"github.com/RokyErickson/doppelganger/pkg/sync"
)

var supportedSessionVersions = []Version{
	Version_Version1,
}

func TestSupportedVersions(t *testing.T) {
	for _, version := range supportedSessionVersions {
		if !version.Supported() {
			t.Error("session version reported as unsupported:", version)
		}
	}
}

func TestDefaultWatchPollingIntervalNonZero(t *testing.T) {
	for _, version := range supportedSessionVersions {
		if version.DefaultWatchPollingInterval() == 0 {
			t.Error("zero-valued default watch polling interval")
		}
	}
}

func TestDefaultFileModeValid(t *testing.T) {
	for _, version := range supportedSessionVersions {
		if err := sync.EnsureDefaultFileModeValid(version.DefaultFileMode()); err != nil {
			t.Error("invalid default file mode:", err)
		}
	}
}

func TestDefaultDirectoryModeValid(t *testing.T) {
	for _, version := range supportedSessionVersions {
		if err := sync.EnsureDefaultDirectoryModeValid(version.DefaultDirectoryMode()); err != nil {
			t.Error("invalid default directory mode:", err)
		}
	}
}
