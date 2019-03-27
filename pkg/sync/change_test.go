package sync

import (
	"testing"
)

func TestChangeCopySlim(t *testing.T) {
	change := &Change{
		Path: "test",
		Old:  nil,
		New:  testDirectory2Entry,
	}

	slim := change.copySlim()

	if err := slim.EnsureValid(); err != nil {
		t.Fatal("slim copy of change is invalid:", err)
	}

	if slim.Path != "test" {
		t.Error("slim copy of change has differing path")
	}

	if !slim.Old.Equal(nil) {
		t.Error("slim copy of change has incorrect old entry")
	}

	if !slim.New.Equal(testEmptyDirectory) {
		t.Error("slim copy of change has incorrect new entry")
	}
}

func TestChangeNilInvalid(t *testing.T) {
	var change *Change
	if change.EnsureValid() == nil {
		t.Error("nil change considered valid")
	}
}

func TestChangeValid(t *testing.T) {
	change := &Change{New: testSymlinkEntry}
	if err := change.EnsureValid(); err != nil {
		t.Error("valid change considered invalid:", err)
	}
}
