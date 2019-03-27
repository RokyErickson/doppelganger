package filesystem

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"unicode/utf8"
)

func TestPathSeparatorSingleByte(t *testing.T) {
	if utf8.RuneLen(os.PathSeparator) != 1 {
		t.Fatal("OS path separator does not have single-byte UTF-8 encoding")
	}
}

func TestDirectoryContentsNotExist(t *testing.T) {
	if _, err := DirectoryContentsByPath("/does/not/exist"); err == nil {
		t.Error("directory listing succeedeed for non-existent path")
	}
}

func TestDirectoryContentsFile(t *testing.T) {

	file, err := ioutil.TempFile("", "doppelganger_filesystem")
	if err != nil {
		t.Fatal("unable to create temporary file:", err)
	} else if err = file.Close(); err != nil {
		t.Error("unable to close temporary file:", err)
	}
	defer os.Remove(file.Name())

	if _, err := DirectoryContentsByPath(file.Name()); err == nil {
		t.Error("directory listing succeedeed for non-directory path")
	}
}

func TestDirectoryContentsGOROOT(t *testing.T) {
	if contents, err := DirectoryContentsByPath(runtime.GOROOT()); err != nil {
		t.Fatal("directory listing failed for GOROOT:", err)
	} else if contents == nil {
		t.Fatal("directory contents nil for GOROOT")
	}
}
