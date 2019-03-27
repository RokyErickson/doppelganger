package sync

import (
	"testing"
)

func changeListsEqual(actualChanges, expectedChanges []*Change) bool {

	if len(actualChanges) != len(expectedChanges) {
		return false
	}

	pathToExpectedChange := make(map[string]*Change, len(expectedChanges))
	for _, expected := range expectedChanges {
		pathToExpectedChange[expected.Path] = expected
	}

	for _, actual := range actualChanges {
		expected, ok := pathToExpectedChange[actual.Path]
		if !ok {
			return false
		}

		if !actual.Old.Equal(expected.Old) {
			return false
		}

		if !actual.New.Equal(expected.New) {
			return false
		}
	}

	return true
}

func conflictListsEqual(actualConflicts, expectedConflicts []*Conflict) bool {
	if len(actualConflicts) != len(expectedConflicts) {
		return false
	}

	pathToExpectedConflict := make(map[string]*Conflict, len(expectedConflicts))
	for _, expected := range expectedConflicts {
		pathToExpectedConflict[expected.Root()] = expected
	}

	for _, actual := range actualConflicts {
		expected, ok := pathToExpectedConflict[actual.Root()]
		if !ok {
			return false
		}

		if !changeListsEqual(actual.AlphaChanges, expected.AlphaChanges) {
			return false
		}

		if !changeListsEqual(actual.BetaChanges, expected.BetaChanges) {
			return false
		}
	}

	return true
}

type reconcileTestCase struct {
	ancestor                *Entry
	alpha                   *Entry
	beta                    *Entry
	synchronizationModes    []SynchronizationMode
	expectedAncestorChanges []*Change
	expectedAlphaChanges    []*Change
	expectedBetaChanges     []*Change
	expectedConflicts       []*Conflict
}

func (c *reconcileTestCase) run(t *testing.T) {
	t.Helper()

	for _, synchronizationMode := range c.synchronizationModes {
		ancestorChanges, alphaChanges, betaChanges, conflicts := Reconcile(
			c.ancestor, c.alpha, c.beta,
			synchronizationMode,
		)

		if !changeListsEqual(ancestorChanges, c.expectedAncestorChanges) {
			t.Error(
				"ancestor changes do not match expected:",
				ancestorChanges, "!=", c.expectedAncestorChanges,
				"using", synchronizationMode,
			)
		}

		if !changeListsEqual(alphaChanges, c.expectedAlphaChanges) {
			t.Error(
				"alpha changes do not match expected:",
				alphaChanges, "!=", c.expectedAlphaChanges,
				"using", synchronizationMode,
			)
		}

		if !changeListsEqual(betaChanges, c.expectedBetaChanges) {
			t.Error(
				"beta changes do not match expected:",
				betaChanges, "!=", c.expectedBetaChanges,
				"using", synchronizationMode,
			)
		}

		if !conflictListsEqual(conflicts, c.expectedConflicts) {
			t.Error(
				"conflicts do not match expected:",
				conflicts, "!=", c.expectedConflicts,
				"using", synchronizationMode,
			)
		}
	}
}

func TestNonDeletionChangesOnly(t *testing.T) {
	changes := []*Change{
		{
			Path: "file",
			New:  testFile1Entry,
		},
		{
			Path: "directory",
			Old:  testDirectory1Entry,
		},
	}
	nonDeletionChanges := nonDeletionChangesOnly(changes)
	if len(nonDeletionChanges) != 1 {
		t.Fatal("more non-deletion changes than expected")
	} else if nonDeletionChanges[0].Path != "file" {
		t.Fatal("non-deletion change has unexpected path")
	}
}

func TestReconcileAllNil(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    nil,
		beta:     nil,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     nil,
		expectedConflicts:       nil,
	}

	testCase.run(t)
}

func TestReconcileDirectoryNothingChanged(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testDirectory1Entry,
		alpha:    testDirectory1Entry,
		beta:     testDirectory1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     nil,
		expectedConflicts:       nil,
	}

	testCase.run(t)
}

func TestReconcileFileNothingChanged(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    testFile1Entry,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     nil,
		expectedConflicts:       nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaModifiedRoot(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    testFile2Entry,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{Old: testFile1Entry, New: testFile2Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileBetaModifiedRootBidirectional(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    testFile1Entry,
		beta:     testFile2Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges: []*Change{
			{Old: testFile1Entry, New: testFile2Entry},
		},
		expectedBetaChanges: nil,
		expectedConflicts:   nil,
	}

	testCase.run(t)
}

func TestReconcileBetaModifiedRootOneWaySafe(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    testFile1Entry,
		beta:     testFile2Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWaySafe,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     nil,
		expectedConflicts: []*Conflict{
			{
				AlphaChanges: []*Change{
					{
						Old: testFile1Entry,
						New: testFile1Entry,
					},
				},
				BetaChanges: []*Change{
					{
						Old: testFile1Entry,
						New: testFile2Entry,
					},
				},
			},
		},
	}

	testCase.run(t)
}

func TestReconcileBetaModifiedRootOneWayReplica(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    testFile1Entry,
		beta:     testFile2Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{Old: testFile2Entry, New: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaDeletedRoot(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    nil,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{Old: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileBetaDeletedRootBidirectional(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    testFile1Entry,
		beta:     nil,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges: []*Change{
			{Old: testFile1Entry},
		},
		expectedBetaChanges: nil,
		expectedConflicts:   nil,
	}

	testCase.run(t)
}

func TestReconcileBetaDeletedRootUnidirectional(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    testFile1Entry,
		beta:     nil,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{New: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileBothDeletedRoot(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    nil,
		beta:     nil,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: []*Change{
			{},
		},
		expectedAlphaChanges: nil,
		expectedBetaChanges:  nil,
		expectedConflicts:    nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaCreatedRoot(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    testFile1Entry,
		beta:     nil,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{New: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileBetaCreatedRootBidirectional(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    nil,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges: []*Change{
			{New: testFile1Entry},
		},
		expectedBetaChanges: nil,
		expectedConflicts:   nil,
	}

	testCase.run(t)
}

func TestReconcileBetaCreatedRootOneWaySafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    nil,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWaySafe,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     nil,
		expectedConflicts:       nil,
	}

	testCase.run(t)
}

func TestReconcileBetaCreatedRootOneWayReplica(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    nil,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{Old: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileBothCreatedSameFile(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    testFile1Entry,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: []*Change{
			{New: testFile1Entry},
		},
		expectedAlphaChanges: nil,
		expectedBetaChanges:  nil,
		expectedConflicts:    nil,
	}

	testCase.run(t)
}

func TestReconcileBothCreatedSameDirectory(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    testDirectory1Entry,
		beta:     testDirectory1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: testDecomposeEntry("", testDirectory1Entry, true),
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     nil,
		expectedConflicts:       nil,
	}

	testCase.run(t)
}

func TestReconcileBothCreatedPartiallyMatchingContentsTwoWaySafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: &Entry{},
		alpha: &Entry{
			Contents: map[string]*Entry{
				"same":      testDirectory1Entry,
				"alpha":     testFile1Entry,
				"different": testFile1Entry,
			},
		},
		beta: &Entry{
			Contents: map[string]*Entry{
				"same":      testDirectory1Entry,
				"beta":      testFile2Entry,
				"different": testDirectory3Entry,
			},
		},
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
		},
		expectedAncestorChanges: testDecomposeEntry("same", testDirectory1Entry, true),
		expectedAlphaChanges: []*Change{
			{Path: "beta", New: testFile2Entry},
		},
		expectedBetaChanges: []*Change{
			{Path: "alpha", New: testFile1Entry},
		},
		expectedConflicts: []*Conflict{
			{
				AlphaChanges: []*Change{
					{
						Path: "different",
						New:  testFile1Entry,
					},
				},
				BetaChanges: []*Change{
					{
						Path: "different",
						New:  testDirectory3Entry,
					},
				},
			},
		},
	}

	testCase.run(t)
}

func TestReconcileBothCreatedPartiallyMatchingContentsTwoWayResolved(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: &Entry{},
		alpha: &Entry{
			Contents: map[string]*Entry{
				"same":      testDirectory1Entry,
				"alpha":     testFile1Entry,
				"different": testFile1Entry,
			},
		},
		beta: &Entry{
			Contents: map[string]*Entry{
				"same":      testDirectory1Entry,
				"beta":      testFile2Entry,
				"different": testDirectory3Entry,
			},
		},
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWayResolved,
		},
		expectedAncestorChanges: testDecomposeEntry("same", testDirectory1Entry, true),
		expectedAlphaChanges: []*Change{
			{Path: "beta", New: testFile2Entry},
		},
		expectedBetaChanges: []*Change{
			{Path: "alpha", New: testFile1Entry},
			{Path: "different", Old: testDirectory3Entry, New: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileBothCreatedPartiallyMatchingContentsOneWaySafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: &Entry{},
		alpha: &Entry{
			Contents: map[string]*Entry{
				"same":      testDirectory1Entry,
				"alpha":     testFile1Entry,
				"different": testFile1Entry,
			},
		},
		beta: &Entry{
			Contents: map[string]*Entry{
				"same":      testDirectory1Entry,
				"beta":      testFile2Entry,
				"different": testDirectory3Entry,
			},
		},
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWaySafe,
		},
		expectedAncestorChanges: testDecomposeEntry("same", testDirectory1Entry, true),
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{Path: "alpha", New: testFile1Entry},
		},
		expectedConflicts: []*Conflict{
			{
				AlphaChanges: []*Change{
					{
						Path: "different",
						New:  testFile1Entry,
					},
				},
				BetaChanges: []*Change{
					{
						Path: "different",
						New:  testDirectory3Entry,
					},
				},
			},
		},
	}

	testCase.run(t)
}

func TestReconcileBothCreatedPartiallyMatchingContentsOneWayReplica(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: &Entry{},
		alpha: &Entry{
			Contents: map[string]*Entry{
				"same":      testDirectory1Entry,
				"alpha":     testFile1Entry,
				"different": testFile1Entry,
			},
		},
		beta: &Entry{
			Contents: map[string]*Entry{
				"same":      testDirectory1Entry,
				"beta":      testFile2Entry,
				"different": testDirectory3Entry,
			},
		},
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: testDecomposeEntry("same", testDirectory1Entry, true),
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{Path: "alpha", New: testFile1Entry},
			{Path: "beta", Old: testFile2Entry},
			{Path: "different", Old: testDirectory3Entry, New: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileBothCreatedDifferentTypesSafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    testDirectory1Entry,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeOneWaySafe,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     nil,
		expectedConflicts: []*Conflict{
			{
				AlphaChanges: []*Change{
					{New: testDirectory1Entry},
				},
				BetaChanges: []*Change{
					{New: testFile1Entry},
				},
			},
		},
	}

	testCase.run(t)
}

func TestReconcileBothCreatedDifferentTypesOverwrite(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: nil,
		alpha:    testDirectory1Entry,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{
				Old: testFile1Entry,
				New: testDirectory1Entry,
			},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaDeletedRootBetaCreatedFileTwoWaySafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testDirectory1Entry,
		alpha:    nil,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges: []*Change{
			{New: testFile1Entry},
		},
		expectedBetaChanges: nil,
		expectedConflicts:   nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaDeletedRootBetaCreatedFileUnsafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testDirectory1Entry,
		alpha:    nil,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{Old: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaDeletedRootBetaCreatedFileOneWaySafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testDirectory1Entry,
		alpha:    nil,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWaySafe,
		},
		expectedAncestorChanges: []*Change{
			{},
		},
		expectedAlphaChanges: nil,
		expectedBetaChanges:  nil,
		expectedConflicts:    nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaCreatedFileBetaDeletedRoot(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testDirectory1Entry,
		alpha:    testFile1Entry,
		beta:     nil,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{New: testFile1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaDeletedRootBetaCreatedDirectoryTwoWaySafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    nil,
		beta:     testDirectory1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges: []*Change{
			{New: testDirectory1Entry},
		},
		expectedBetaChanges: nil,
		expectedConflicts:   nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaDeletedRootBetaCreatedDirectoryUnsafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    nil,
		beta:     testDirectory1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{Old: testDirectory1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaDeletedRootBetaCreatedDirectoryOneWaySafe(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    nil,
		beta:     testDirectory1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWaySafe,
		},
		expectedAncestorChanges: []*Change{
			{},
		},
		expectedAlphaChanges: nil,
		expectedBetaChanges:  nil,
		expectedConflicts:    nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaCreatedDirectoryBetaDeletedRootNonBetaWinsAll(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testFile1Entry,
		alpha:    testDirectory1Entry,
		beta:     nil,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{New: testDirectory1Entry},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaPartiallyDeletedDirectory(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testDirectory2Entry,
		alpha:    testDirectory3Entry,
		beta:     testDirectory2Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     diff("", testDirectory2Entry, testDirectory3Entry),
		expectedConflicts:       nil,
	}

	testCase.run(t)
}

func TestReconcileBetaPartiallyDeletedDirectoryBidirectional(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testDirectory2Entry,
		alpha:    testDirectory2Entry,
		beta:     testDirectory3Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    diff("", testDirectory2Entry, testDirectory3Entry),
		expectedBetaChanges:     nil,
		expectedConflicts:       nil,
	}

	testCase.run(t)
}

func TestReconcileBetaPartiallyDeletedDirectoryUnidirectional(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testDirectory2Entry,
		alpha:    testDirectory2Entry,
		beta:     testDirectory3Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     diff("", testDirectory3Entry, testDirectory2Entry),
		expectedConflicts:       nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaReplacedDirectoryBetaPartiallyDeletedDirectory(t *testing.T) {

	testCase := reconcileTestCase{
		ancestor: testDirectory2Entry,
		alpha:    testFile1Entry,
		beta:     testDirectory3Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWaySafe,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{
				Old: testDirectory3Entry,
				New: testFile1Entry,
			},
		},
		expectedConflicts: nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaPartiallyDeletedDirectoryBetaReplacedDirectoryTwoWaySafe(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testDirectory2Entry,
		alpha:    testDirectory3Entry,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWaySafe,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges: []*Change{
			{
				Old: testDirectory3Entry,
				New: testFile1Entry,
			},
		},
		expectedBetaChanges: nil,
		expectedConflicts:   nil,
	}

	testCase.run(t)
}

func TestReconcileAlphaPartiallyDeletedDirectoryBetaReplacedDirectoryUnsafe(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testDirectory2Entry,
		alpha:    testDirectory3Entry,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeTwoWayResolved,
			SynchronizationMode_SynchronizationModeOneWayReplica,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges: []*Change{
			{
				Old: testFile1Entry,
				New: testDirectory3Entry,
			},
		},
		expectedConflicts: nil,
	}
	testCase.run(t)
}

func TestReconcileAlphaPartiallyDeletedDirectoryBetaReplacedDirectoryOneWaySafe(t *testing.T) {
	testCase := reconcileTestCase{
		ancestor: testDirectory2Entry,
		alpha:    testDirectory3Entry,
		beta:     testFile1Entry,
		synchronizationModes: []SynchronizationMode{
			SynchronizationMode_SynchronizationModeOneWaySafe,
		},
		expectedAncestorChanges: nil,
		expectedAlphaChanges:    nil,
		expectedBetaChanges:     nil,
		expectedConflicts: []*Conflict{
			{
				AlphaChanges: diff("", testDirectory2Entry, testDirectory3Entry),
				BetaChanges:  diff("", testDirectory2Entry, testFile1Entry),
			},
		},
	}

	testCase.run(t)
}
