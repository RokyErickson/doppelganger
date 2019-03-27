package sync

import (
	"testing"
)

func stripExecutabilityRecursive(snapshot *Entry) {

	if snapshot == nil {
		return
	}

	if snapshot.Kind == EntryKind_Directory {
		for _, entry := range snapshot.Contents {
			stripExecutabilityRecursive(entry)
		}
	} else if snapshot.Kind == EntryKind_File {
		snapshot.Executable = false
	}
}

func stripExecutability(snapshot *Entry) *Entry {

	result := snapshot.Copy()

	stripExecutabilityRecursive(result)

	return result
}

func TestExecutabilityPropagateNil(t *testing.T) {
	if PropagateExecutability(testDirectory1Entry, testDirectory1Entry, nil) != nil {
		t.Fatal("executability propagation to nil entry did not return nil")
	}
}

func TestExecutabilityPropagationCycle(t *testing.T) {

	stripped := stripExecutability(testDirectory1Entry)
	if stripped == testDirectory1Entry {
		t.Fatal("executability stripping did not make entry copy")
	} else if stripped.Equal(testDirectory1Entry) {
		t.Fatal("stripped directory entry considered equal to original")
	}

	fixed := PropagateExecutability(nil, nil, stripped)
	if fixed == stripped {
		t.Fatal("executability propagation did not make entry copy")
	} else if !fixed.Equal(stripped) {
		t.Fatal("executability propagation from nil ancestor/source made changes to entry")
	}

	fixed = PropagateExecutability(testDirectory1Entry, nil, stripped)
	if fixed == stripped {
		t.Fatal("executability propagation did not make entry copy")
	} else if !fixed.Equal(testDirectory1Entry) {
		t.Fatal("executability propagation from ancestor incorrect")
	}

	fixed = PropagateExecutability(nil, testDirectory1Entry, stripped)
	if fixed == stripped {
		t.Fatal("executability propagation did not make entry copy")
	} else if !fixed.Equal(testDirectory1Entry) {
		t.Fatal("executability propagation from source incorrect")
	}
}
