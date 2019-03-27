package sync

import (
	"bytes"

	"github.com/pkg/errors"
)

type stagingPathFinder struct {
	paths []string

	digests [][]byte
}

func (f *stagingPathFinder) find(path string, entry *Entry) error {

	if entry == nil {
		return nil
	}

	if entry.Kind == EntryKind_Directory {
		for name, entry := range entry.Contents {
			if err := f.find(pathJoin(path, name), entry); err != nil {
				return err
			}
		}
	} else if entry.Kind == EntryKind_File {
		f.paths = append(f.paths, path)
		f.digests = append(f.digests, entry.Digest)
	} else if entry.Kind == EntryKind_Symlink {
		return nil
	} else {
		return errors.New("unknown entry type encountered")
	}

	return nil
}

func TransitionDependencies(transitions []*Change) ([]string, [][]byte, error) {
	finder := &stagingPathFinder{}
	for _, t := range transitions {
		fileToFileSameContents := t.Old != nil && t.New != nil &&
			t.Old.Kind == EntryKind_File && t.New.Kind == EntryKind_File &&
			bytes.Equal(t.Old.Digest, t.New.Digest)
		if fileToFileSameContents {
			continue
		}

		if err := finder.find(t.Path, t.New); err != nil {
			return nil, nil, errors.Wrap(err, "unable to find staging paths")
		}
	}
	return finder.paths, finder.digests, nil
}
