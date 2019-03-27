package filesystem

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicNonExistentDirectory(t *testing.T) {
	if WriteFileAtomic("/does/not/exist", []byte{}, 0600) == nil {
		t.Error("atomic file write did not fail for non-existent path")
	}
}

func TestWriteFileAtomic(t *testing.T) {

	directory, err := ioutil.TempDir("", "doppelganger_write_file_atomic")
	if err != nil {
		t.Fatal("unable to create temporary directory:", err)
	}
	defer os.RemoveAll(directory)

	target := filepath.Join(directory, "file")

	contents := []byte{0, 1, 2, 3, 4, 5, 6}

	if err := WriteFileAtomic(target, contents, 0600); err != nil {
		t.Fatal("atomic file write failed:", err)
	}

	if data, err := ioutil.ReadFile(target); err != nil {
		t.Fatal("unable to read back file:", err)
	} else if !bytes.Equal(data, contents) {
		t.Error("file contents did not match expected")
	}
}
