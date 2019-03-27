package filesystem

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestMarkHidden(t *testing.T) {

	hiddenFile, err := ioutil.TempFile("", ".doppelganger_filesystem_hidden")
	if err != nil {
		t.Fatal("unable to create temporary hiddenFile file:", err)
	}
	hiddenFile.Close()
	defer os.Remove(hiddenFile.Name())

	if err := markHidden(hiddenFile.Name()); err != nil {
		t.Fatal("unable to mark file as hidden")
	}

}
