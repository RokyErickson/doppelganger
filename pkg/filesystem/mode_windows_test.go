package filesystem

import (
	"os"
	"testing"
)

func TestModePermissionsMaskMatchesOS(t *testing.T) {
	if ModePermissionsMask != Mode(os.ModePerm) {
		t.Error("ModePermissionsMask does not match expected value")
	}
}
