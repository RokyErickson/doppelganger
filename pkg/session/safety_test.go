package session

import (
	"testing"
)

func TestFilteredPathsAreSubset(t *testing.T) {
	testCases := []struct {
		filteredPaths []string
		originalPaths []string
		expected      bool
	}{
		{nil, nil, true},
		{nil, []string{}, true},
		{[]string{}, []string{}, true},
		{[]string{}, nil, true},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a"}, []string{"a", "b"}, true},
		{[]string{"b"}, []string{"a", "b"}, true},
		{[]string{"c"}, nil, false},
		{[]string{"c"}, []string{}, false},
		{[]string{"c"}, []string{"a"}, false},
		{[]string{"c"}, []string{"a", "b"}, false},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"b", "a"}, []string{"a", "b"}, false},
	}

	for c, testCase := range testCases {
		if result := filteredPathsAreSubset(
			testCase.filteredPaths,
			testCase.originalPaths,
		); result != testCase.expected {
			t.Errorf(
				"result did not match expected for test case %d: %t != %t",
				c,
				result,
				testCase.expected,
			)
		}
	}
}
