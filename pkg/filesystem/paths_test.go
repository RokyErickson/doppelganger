package filesystem

import (
	"os"
	"testing"
)

const (
	testingDirectoryName = "testing"
)

func TestDoppelganger(t *testing.T) {

	path, err := Doppelganger(true, testingDirectoryName)
	if err != nil {
		t.Fatal("unable to create testing subdirectory:", err)
	}
	defer os.RemoveAll(path)

	if info, err := os.Lstat(path); err != nil {
		t.Fatal("unable to probe testing subdirectory:", err)
	} else if !info.IsDir() {
		t.Error("Doppelganger subpath is not a directory")
	}
}
