package session

import (
	"github.com/RokyErickson/doppelganger/pkg/sync"
)

func isRootDeletion(change *sync.Change) bool {
	return change.Path == "" && change.Old != nil && change.New == nil
}

func isRootTypeChange(change *sync.Change) bool {
	return change.Path == "" &&
		change.Old != nil && change.New != nil &&
		change.Old.Kind != change.New.Kind
}

func filteredPathsAreSubset(filteredPaths, originalPaths []string) bool {

	for _, filtered := range filteredPaths {
		matched := false

		for o, original := range originalPaths {
			if original == filtered {
				originalPaths = originalPaths[o+1:]
				matched = true
				break
			}
		}

		if !matched {
			return false
		}
	}

	return true
}
