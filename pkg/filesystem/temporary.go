package filesystem

import (
	"strings"
)

const (
	TemporaryNamePrefix = ".doppelganger-temporary-"
)

func IsTemporaryFileName(name string) bool {
	return strings.HasPrefix(name, TemporaryNamePrefix)
}
