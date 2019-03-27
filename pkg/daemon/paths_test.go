package daemon

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSubpath(t *testing.T) {

	path, err := subpath("something")
	if err != nil {
		t.Fatal("unable to compute subpath:", err)
	}

	if s, err := os.Lstat(filepath.Dir(path)); err != nil {
		t.Fatal("unable to verify that daemon subdirectory exists:", err)
	} else if !s.IsDir() {
		t.Error("daemon subdirectory is not a directory")
	}
}
