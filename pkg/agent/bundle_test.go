package agent

import (
	"os"
	"runtime"
	"testing"

	"github.com/pkg/errors"
)

func init() {
	if err := CopyBundleForTesting(); err != nil {
		panic(errors.Wrap(err, "unable to copy agent bundle for testing"))
	}
}

func TestExecutableForInvalidOS(t *testing.T) {
	if _, err := executableForPlatform("fakeos", runtime.GOARCH); err == nil {
		t.Fatal("extracting agent executable succeeded for invalid OS")
	}
}

func TestExecutableForInvalidArchitecture(t *testing.T) {
	if _, err := executableForPlatform(runtime.GOOS, "fakearch"); err == nil {
		t.Fatal("extracting agent executable succeeded for invalid architecture")
	}
}

func TestExecutableForInvalidPair(t *testing.T) {
	if _, err := executableForPlatform("fakeos", "fakearch"); err == nil {
		t.Fatal("extracting agent executable succeeded for invalid architecture")
	}
}

func TestExecutableForPlatform(t *testing.T) {
	if executable, err := executableForPlatform(runtime.GOOS, runtime.GOARCH); err != nil {
		t.Fatal("unable to extract agent bundle for current platform:", err)
	} else if err = os.Remove(executable); err != nil {
		t.Error("unable to remove agent executable:", err)
	}
}
