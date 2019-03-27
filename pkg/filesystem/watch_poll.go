package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultInitialContentMapCapacity = 1024
)

func fileInfoEqual(first, second os.FileInfo) bool {

	if first.Mode() != second.Mode() {
		return false
	}

	if first.IsDir() {
		return true
	}

	return first.Size() == second.Size() &&
		first.ModTime().Equal(second.ModTime())
}

func poll(root string, existing map[string]os.FileInfo, trackChanges bool) (map[string]os.FileInfo, bool, map[string]bool, error) {

	initialContentMapCapacity := len(existing)
	if initialContentMapCapacity == 0 {
		initialContentMapCapacity = defaultInitialContentMapCapacity
	}
	contents := make(map[string]os.FileInfo, initialContentMapCapacity)

	var changes map[string]bool
	if trackChanges {

		changes = make(map[string]bool)
	}

	changed := false
	rootDoesNotExist := false
	visitor := func(path string, info os.FileInfo, err error) error {

		if err != nil {

			if path == root && os.IsNotExist(err) {
				changed = len(existing) > 0
				rootDoesNotExist = true
				return err
			}

			if os.IsNotExist(err) {
				return nil
			}

			return err
		}

		if IsTemporaryFileName(filepath.Base(path)) {
			return nil
		}

		contents[path] = info

		pathChanged := false
		if previous, ok := existing[path]; !ok || !fileInfoEqual(info, previous) {
			pathChanged = true
		}

		if pathChanged {
			changed = true
		}

		if trackChanges && pathChanged {
			if info.IsDir() {
				changes[path] = true
			}
			if path != root {
				changes[filepath.Dir(path)] = true
			}
		}

		return nil
	}

	if err := Walk(root, visitor); err != nil && !rootDoesNotExist {
		return nil, false, nil, errors.Wrap(err, "unable to perform filesystem walk")
	}

	if len(contents) != len(existing) {
		changed = true
	}

	return contents, changed, changes, nil
}

func watchPoll(context context.Context, root string, events chan struct{}, pollInterval uint32) error {

	if pollInterval == 0 {
		return errors.New("polling interval must be greater than 0 seconds")
	}
	pollIntervalDuration := time.Duration(pollInterval) * time.Second

	timer := time.NewTimer(0)

	var contents map[string]os.FileInfo
	for {
		select {
		case <-context.Done():

			return errors.New("watch cancelled")
		case <-timer.C:

			newContents, changed, _, err := poll(root, contents, false)
			if err != nil || !changed {
				timer.Reset(pollIntervalDuration)
				continue
			}

			contents = newContents

			select {
			case events <- struct{}{}:
			default:
			}

			timer.Reset(pollIntervalDuration)
		}
	}
}
