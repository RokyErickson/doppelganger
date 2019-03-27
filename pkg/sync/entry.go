package sync

import (
	"bytes"
	"strings"

	"github.com/pkg/errors"
)

func (e *Entry) EnsureValid() error {

	if e == nil {
		return nil
	}
	if e.Kind == EntryKind_Directory {

		if e.Executable {
			return errors.New("executable directory detected")
		} else if e.Digest != nil {
			return errors.New("non-nil directory digest detected")
		} else if e.Target != "" {
			return errors.New("non-empty symlink target detected for directory")
		}

		for name, entry := range e.Contents {
			if name == "" {
				return errors.New("empty content name detected")
			} else if strings.IndexByte(name, '/') != -1 {
				return errors.New("content name contains path separator")
			} else if entry == nil {
				return errors.New("nil content detected")
			} else if err := entry.EnsureValid(); err != nil {
				return err
			}
		}
	} else if e.Kind == EntryKind_File {

		if e.Contents != nil {
			return errors.New("non-nil file contents detected")
		} else if e.Target != "" {
			return errors.New("non-empty symlink target detected for file")
		}

		if len(e.Digest) == 0 {
			return errors.New("file with empty digest detected")
		}
	} else if e.Kind == EntryKind_Symlink {

		if e.Executable {
			return errors.New("executable symlink detected")
		} else if e.Digest != nil {
			return errors.New("non-nil symlink digest detected")
		} else if e.Contents != nil {
			return errors.New("non-nil symlink contents detected")
		}

		if e.Target == "" {
			return errors.New("symlink with empty target detected")
		}

	} else {
		return errors.New("unknown entry kind detected")
	}

	return nil
}

func (e *Entry) Count() uint64 {

	if e == nil {
		return 0
	}

	result := uint64(1)

	if e.Kind == EntryKind_Directory {
		for _, entry := range e.Contents {

			result += entry.Count()
		}
	}

	return result
}

func (e *Entry) equalShallow(other *Entry) bool {

	if e == nil && other == nil {
		return true
	}

	if e == nil || other == nil {
		return false
	}

	return e.Kind == other.Kind &&
		e.Executable == other.Executable &&
		bytes.Equal(e.Digest, other.Digest) &&
		e.Target == other.Target
}

func (e *Entry) Equal(other *Entry) bool {
	if !e.equalShallow(other) {
		return false
	}

	if e == nil && other == nil {
		return true
	}

	if len(e.Contents) != len(other.Contents) {
		return false
	}
	for name, entry := range e.Contents {
		otherEntry, ok := other.Contents[name]
		if !ok || !entry.Equal(otherEntry) {
			return false
		}
	}

	return true
}

func (e *Entry) copySlim() *Entry {
	if e == nil {
		return nil
	}

	return &Entry{
		Kind:       e.Kind,
		Executable: e.Executable,
		Digest:     e.Digest,
		Target:     e.Target,
	}
}

func (e *Entry) Copy() *Entry {
	if e == nil {
		return nil
	}

	result := &Entry{
		Kind:       e.Kind,
		Executable: e.Executable,
		Digest:     e.Digest,
		Target:     e.Target,
	}

	if len(e.Contents) == 0 {
		return result
	}

	result.Contents = make(map[string]*Entry)
	for name, entry := range e.Contents {
		result.Contents[name] = entry.Copy()
	}

	return result
}
