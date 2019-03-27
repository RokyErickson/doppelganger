package sync

import (
	"testing"
)

func TestApplyRootSwap(t *testing.T) {
	changes := []*Change{
		{
			Old: testDirectory1Entry,
			New: testFile1Entry,
		},
	}

	if result, err := Apply(testDirectory1Entry, changes); err != nil {
		t.Fatal("unable to apply changes:", err)
	} else if !result.Equal(testFile1Entry) {
		t.Error("mismatch after root replacement")
	}
}

func TestApplyDiff(t *testing.T) {
	changes := diff("", testDirectory1Entry, testDirectory2Entry)

	if result, err := Apply(testDirectory1Entry, changes); err != nil {
		t.Fatal("unable to apply changes:", err)
	} else if !result.Equal(testDirectory2Entry) {
		t.Error("mismatch after diff/apply cycle")
	}
}

func TestApplyMissingParentPath(t *testing.T) {
	changes := []*Change{
		{
			Path: "this/does/not/exist",
			New:  testFile1Entry,
		},
	}

	if _, err := Apply(testDirectory1Entry, changes); err == nil {
		t.Fatal("change referencing invalid path did not fail to apply")
	}
}
