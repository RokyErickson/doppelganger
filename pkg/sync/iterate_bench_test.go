package sync

import (
	"testing"
)

func BenchmarkThreeWayNameUnionUnaltered(b *testing.B) {
	contents := map[string]*Entry{
		"first":   nil,
		"second":  nil,
		"third":   nil,
		"fourth":  nil,
		"fifth":   nil,
		"sixth":   nil,
		"seventh": nil,
		"eighth":  nil,
		"ninth":   nil,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nameUnion(contents, contents, contents)
	}
}

func BenchmarkThreeWayNameUnionOneAltered(b *testing.B) {

	ancestor := map[string]*Entry{
		"first":   nil,
		"second":  nil,
		"third":   nil,
		"fourth":  nil,
		"fifth":   nil,
		"sixth":   nil,
		"seventh": nil,
		"eighth":  nil,
		"ninth":   nil,
	}
	altered := map[string]*Entry{
		"first":   nil,
		"second":  nil,
		"third":   nil,
		"fourth":  nil,
		"fifth":   nil,
		"sixth":   nil,
		"seventh": nil,
		"eighth":  nil,
		"ninth":   nil,
		"tenth":   nil,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nameUnion(ancestor, ancestor, altered)
	}
}
