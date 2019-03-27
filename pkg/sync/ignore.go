package sync

import (
	pathpkg "path"
	"strings"

	"github.com/pkg/errors"

	"github.com/bmatcuk/doublestar"
)

func (m IgnoreVCSMode) IsDefault() bool {
	return m == IgnoreVCSMode_IgnoreVCSDefault
}

func (m *IgnoreVCSMode) UnmarshalText(textBytes []byte) error {

	text := string(textBytes)

	switch text {
	case "true":
		*m = IgnoreVCSMode_IgnoreVCS
	case "false":
		*m = IgnoreVCSMode_PropagateVCS
	default:
		return errors.Errorf("unknown VCS ignore specification: %s", text)
	}

	return nil
}

func (m IgnoreVCSMode) Supported() bool {
	switch m {
	case IgnoreVCSMode_IgnoreVCS:
		return true
	case IgnoreVCSMode_PropagateVCS:
		return true
	default:
		return false
	}
}

func (m IgnoreVCSMode) Description() string {
	switch m {
	case IgnoreVCSMode_IgnoreVCSDefault:
		return "Default"
	case IgnoreVCSMode_IgnoreVCS:
		return "Ignore"
	case IgnoreVCSMode_PropagateVCS:
		return "Propagate"
	default:
		return "Unknown"
	}
}

var DefaultVCSIgnores = []string{
	".git/",
	".svn/",
	".hg/",
	".bzr/",
	"_darcs/",
}

type ignorePattern struct {
	negated       bool
	directoryOnly bool
	matchLeaf     bool
	pattern       string
}

func newIgnorePattern(pattern string) (*ignorePattern, error) {
	if pattern == "" || pattern == "!" {
		return nil, errors.New("empty pattern")
	} else if pattern == "/" || pattern == "!/" {
		return nil, errors.New("root pattern")
	} else if pattern == "//" || pattern == "!//" {
		return nil, errors.New("root directory pattern")
	}

	negated := false
	if pattern[0] == '!' {
		negated = true
		pattern = pattern[1:]
	}

	absolute := false
	if pattern[0] == '/' {
		absolute = true
		pattern = pattern[1:]
	}

	directoryOnly := false
	if pattern[len(pattern)-1] == '/' {
		directoryOnly = true
		pattern = pattern[:len(pattern)-1]
	}

	containsSlash := strings.IndexByte(pattern, '/') >= 0

	if _, err := doublestar.Match(pattern, "a"); err != nil {
		return nil, errors.Wrap(err, "unable to validate pattern")
	}

	return &ignorePattern{
		negated:       negated,
		directoryOnly: directoryOnly,
		matchLeaf:     (!absolute && !containsSlash),
		pattern:       pattern,
	}, nil
}

func (i *ignorePattern) matches(path string, directory bool) (bool, bool) {
	if i.directoryOnly && !directory {
		return false, false
	}

	if match, _ := doublestar.Match(i.pattern, path); match {
		return true, i.negated
	}

	if i.matchLeaf && path != "" {
		if match, _ := doublestar.Match(i.pattern, pathpkg.Base(path)); match {
			return true, i.negated
		}
	}

	return false, false
}

func ValidIgnorePattern(pattern string) bool {
	_, err := newIgnorePattern(pattern)
	return err == nil
}

type ignorer struct {
	patterns []*ignorePattern
}

func newIgnorer(patterns []string) (*ignorer, error) {
	ignorePatterns := make([]*ignorePattern, len(patterns))
	for i, p := range patterns {
		if ip, err := newIgnorePattern(p); err != nil {
			return nil, errors.Wrap(err, "unable to parse pattern")
		} else {
			ignorePatterns[i] = ip
		}
	}

	return &ignorer{ignorePatterns}, nil
}

func (i *ignorer) ignored(path string, directory bool) bool {

	ignored := false

	for _, p := range i.patterns {
		if match, negated := p.matches(path, directory); !match {
			continue
		} else {
			ignored = !negated
		}
	}

	return ignored
}

type IgnoreCacheKey struct {
	path      string
	directory bool
}

type IgnoreCache map[IgnoreCacheKey]bool
