package sync

import (
	"strings"

	"github.com/pkg/errors"
)

func Apply(base *Entry, changes []*Change) (*Entry, error) {
	result := base.Copy()

	for _, c := range changes {
		if c.Path == "" {
			result = c.New
			continue
		}

		parent := result
		components := strings.Split(c.Path, "/")
		for len(components) > 1 {
			child, ok := parent.Contents[components[0]]
			if !ok {
				return nil, errors.New("unable to resolve parent path")
			}
			parent = child
			components = components[1:]
		}

		if c.New == nil {
			delete(parent.Contents, components[0])
		} else {
			if parent.Contents == nil {
				parent.Contents = make(map[string]*Entry)
			}
			parent.Contents[components[0]] = c.New
		}
	}

	return result, nil
}
