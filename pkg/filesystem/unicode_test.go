package filesystem

import (
	"os"
	"runtime"
	"testing"
)

func TestDecomposesUnicodeByPathDarwinHFS(t *testing.T) {

	if runtime.GOOS != "darwin" {
		t.Skip()
	}

	hfsRoot := os.Getenv("DOPPELGANGER_TEST_HFS_ROOT")
	if hfsRoot == "" {
		t.Skip()
	}

	if decomposes, err := DecomposesUnicodeByPath(hfsRoot); err != nil {
		t.Fatal("unable to probe Unicode decomposition:", err)
	} else if !decomposes {
		t.Error("Unicode decomposition behavior does not match expected")
	}
}

func TestDecomposesUnicodeByPathDarwinAPFS(t *testing.T) {

	apfsRoot := os.Getenv("DOPPELGANGER_TEST_APFS_ROOT")
	if apfsRoot == "" {
		t.Skip()
	}

	if decomposes, err := DecomposesUnicodeByPath(apfsRoot); err != nil {
		t.Fatal("unable to probe Unicode decomposition:", err)
	} else if decomposes {
		t.Error("Unicode decomposition behavior does not match expected")
	}
}

func TestDecomposesUnicodeByPathOSPartition(t *testing.T) {

	if runtime.GOOS == "darwin" {
		t.Skip()
	}

	if decomposes, err := DecomposesUnicodeByPath("."); err != nil {
		t.Fatal("unable to probe Unicode decomposition:", err)
	} else if decomposes {
		t.Error("Unicode decomposition behavior does not match expected")
	}
}

type decomposesUnicodeTestCase struct {
	path     string
	expected bool
}

func (c *decomposesUnicodeTestCase) run(t *testing.T) {

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

	if decomposes, err := DecomposesUnicode(directory); err != nil {
		t.Fatal("unable to probe Unicode decomposition:", err)
	} else if decomposes != c.expected {
		t.Error("Unicode decomposition behavior does not match expected")
	}
}

func TestDecomposesUnicodeDarwinHFS(t *testing.T) {

	if runtime.GOOS != "darwin" {
		t.Skip()
	}

	hfsRoot := os.Getenv("DOPPELGANGER_TEST_HFS_ROOT")
	if hfsRoot == "" {
		t.Skip()
	}

	testCase := &decomposesUnicodeTestCase{
		path:     hfsRoot,
		expected: true,
	}

	testCase.run(t)
}

func TestDecomposesUnicodeDarwinAPFS(t *testing.T) {

	apfsRoot := os.Getenv("DOPPELGANGER_TEST_APFS_ROOT")
	if apfsRoot == "" {
		t.Skip()
	}

	testCase := &decomposesUnicodeTestCase{
		path:     apfsRoot,
		expected: false,
	}

	testCase.run(t)
}

func TestDecomposesUnicodeHomeDirectory(t *testing.T) {

	if runtime.GOOS == "darwin" {
		t.Skip()
	}

	testCase := &decomposesUnicodeTestCase{
		path:     HomeDirectory,
		expected: false,
	}

	testCase.run(t)
}
