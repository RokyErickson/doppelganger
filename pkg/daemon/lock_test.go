package daemon

import (
	"bytes"
	"strings"
	"testing"
	"github.com/polydawn/gosh"
	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
)

const (
	lockTestExecutablePackage = "github.com/RokyErickson/doppelganger/pkg/daemon/locktest"
	lockTestFailMessage       = "Doppelganger lock acquisition failed"
)

func TestLockCycle(t *testing.T) {

	lock, err := AcquireLock()
	if err != nil {
		t.Fatal("unable to acquire lock:", err)
	}

	if err := lock.Unlock(); err != nil {
		t.Fatal("unable to release lock:", err)
	}
}

func TestLockDuplicateFail(t *testing.T) {

	doppelgangerSourcePath, err := doppelganger.SourceTreePath()
	if err != nil {
		t.Fatal("unable to compute path to Doppelganger source tree:", err)
	}

	lock, err := AcquireLock()
	if err != nil {
		t.Fatal("unable to acquire lock:", err)
	}
	defer lock.Unlock()

	testCommand := gosh.Gosh("go", "run", lockTestExecutablePackage, gosh.Opts{Cwd: doppelgangerSourcePath})
	testCommand.Run()
	}
}
