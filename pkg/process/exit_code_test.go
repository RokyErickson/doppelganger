package process

import (
	"github.com/polydawn/gosh"
	"runtime"
	"testing"
)

func TestExitCode(t *testing.T) {

	command := gosh.Gosh("go", "doppelganger-test-invalid")
	command.Run()
}

func TestIsPOSIXShellInvalidCommand(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip()
	}

	command := gosh.Gosh("/bin/sh", "-c", "/dev/null")
	command.Run()
}

func TestIsPOSIXShellCommandNotFound(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}

	command := gosh.Gosh("/bin/sh", "doppelganger-test-not-exist")
	command.Run()
}
