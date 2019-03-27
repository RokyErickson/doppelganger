package filesystem

import (
	"os"
	"runtime"
	"testing"
)

type preservesExecutabilityByPathTestCase struct {
	path string

	expected bool
}

func (c *preservesExecutabilityByPathTestCase) run(t *testing.T) {

	t.Helper()

	if preserves, err := PreservesExecutabilityByPath(c.path); err != nil {
		t.Fatal("unable to probe executability preservation:", err)
	} else if preserves != c.expected {
		t.Error("executability preservation behavior does not match expected")
	}
}

func TestPreservesExecutabilityByPathHomeDirectory(t *testing.T) {

	testCase := &preservesExecutabilityByPathTestCase{
		path:     HomeDirectory,
		expected: runtime.GOOS != "windows",
	}

	testCase.run(t)
}

func TestPreservesExecutabilityByPathFAT32(t *testing.T) {

	fat32Root := os.Getenv("DOPPELGANGER_TEST_FAT32_ROOT")
	if fat32Root == "" {
		t.Skip()
	}

	testCase := &preservesExecutabilityByPathTestCase{
		path:     fat32Root,
		expected: false,
	}

	testCase.run(t)
}

type preservesExecutabilityTestCase struct {
	path string

	expected bool
}

func (c *preservesExecutabilityTestCase) run(t *testing.T) {

	t.Helper()

	object, metadata, err := Open(c.path, false)
	var directory *Directory
	var ok bool
	if err != nil {
		t.Fatal("unable to open path:", err)
	} else if metadata.Mode&ModeTypeMask != ModeTypeDirectory {
		t.Fatal("path is not a directory")
	} else if directory, ok = object.(*Directory); !ok {
		t.Fatal("filesystem object did not convert to directory")
	}
	defer directory.Close()

	if preserves, err := PreservesExecutability(directory); err != nil {
		t.Fatal("unable to probe executability preservation:", err)
	} else if preserves != c.expected {
		t.Error("executability preservation behavior does not match expected")
	}
}

func TestPreservesExecutabilityHomeDirectory(t *testing.T) {

	testCase := &preservesExecutabilityTestCase{
		path:     HomeDirectory,
		expected: runtime.GOOS != "windows",
	}

	testCase.run(t)
}

func TestPreservesExecutabilityFAT32(t *testing.T) {

	fat32Root := os.Getenv("DOPPELGANGER_TEST_FAT32_ROOT")
	if fat32Root == "" {
		t.Skip()
	}

	testCase := &preservesExecutabilityTestCase{
		path:     fat32Root,
		expected: false,
	}

	testCase.run(t)
}
