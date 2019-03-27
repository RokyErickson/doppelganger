// +build windows darwin,cgo

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pkg/errors"
)

const (
	watchRootParameterPollingInterval = 1 * time.Second
	watchRestartWait                  = 1 * time.Second
)

func isParentOrSelf(parent, child string) bool {
	parentLength := len(parent)
	childLength := len(child)
	if childLength < parentLength {
		return false
	} else if parent != child[:parentLength] {
		return false
	} else if childLength > parentLength {
		return os.IsPathSeparator(child[parentLength])
	}
	return true
}

func watchNative(context context.Context, root string, events chan struct{}, _ uint32) error {

	var watchRoot string
	if runtime.GOOS == "darwin" {
		watchRoot = root
	} else if runtime.GOOS == "windows" {
		watchRoot = filepath.Dir(root)
	} else {
		panic("unhandled platform case")
	}

	var watchRootExists bool
	var watchRootMetadata os.FileInfo
	var forceRecreate bool

	dummyEventPaths := make(chan string)
	eventPaths := dummyEventPaths

	var watch *recursiveWatch

	coalescingTimer := time.NewTimer(watchNativeCoalescingWindow)
	if !coalescingTimer.Stop() {
		<-coalescingTimer.C
	}

	watchRootCheckTimer := time.NewTimer(0)

	defer func() {

		if watch != nil {
			watch.stop()
		}

		coalescingTimer.Stop()

		watchRootCheckTimer.Stop()
	}()

	for {
		select {
		case <-context.Done():

			return errors.New("watch cancelled")
		case path, ok := <-eventPaths:

			if !ok {

				watch.stop()
				watch = nil
				eventPaths = dummyEventPaths

				forceRecreate = true

				if !watchRootCheckTimer.Stop() {
					<-watchRootCheckTimer.C
				}
				watchRootCheckTimer.Reset(watchRestartWait)

				continue
			}

			if IsTemporaryFileName(filepath.Base(path)) {

				continue
			} else if runtime.GOOS == "windows" && !isParentOrSelf(root, path) {

				continue
			} else {

				if !coalescingTimer.Stop() {
					select {
					case <-coalescingTimer.C:
					default:
					}
				}
				coalescingTimer.Reset(watchNativeCoalescingWindow)
			}
		case <-coalescingTimer.C:

			select {
			case events <- struct{}{}:
			default:
			}
		case <-watchRootCheckTimer.C:

			var watchRootCurrentlyExists bool
			var currentWatchRootMetadata os.FileInfo
			if m, err := os.Lstat(watchRoot); err != nil {
				if !os.IsNotExist(err) {
					return errors.Wrap(err, "unable to probe root metadata")
				}
			} else {
				watchRootCurrentlyExists = true
				currentWatchRootMetadata = m
			}

			recreate := forceRecreate ||
				watchRootCurrentlyExists != watchRootExists ||
				!watchRootParametersEqual(currentWatchRootMetadata, watchRootMetadata)

			if recreate && runtime.GOOS == "darwin" {
				overrideRecreate := !forceRecreate &&
					watchRootExists && watchRootCurrentlyExists &&
					currentWatchRootMetadata.Mode() == watchRootMetadata.Mode() &&
					currentWatchRootMetadata.Mode()&os.ModeType == 0
				if overrideRecreate {
					recreate = false
				}
			}

			forceRecreate = false

			if recreate {

				if watch != nil {
					watch.stop()
					watch = nil
					eventPaths = dummyEventPaths
				}

				if watchRootCurrentlyExists {
					if w, err := newRecursiveWatch(watchRoot, currentWatchRootMetadata); err != nil {
						forceRecreate = true
						watchRootCheckTimer.Reset(watchRestartWait)
						continue
					} else {
						watch = w
						eventPaths = w.eventPaths
					}
				}

				select {
				case events <- struct{}{}:
				default:
				}
			}

			watchRootExists = watchRootCurrentlyExists
			watchRootMetadata = currentWatchRootMetadata

			watchRootCheckTimer.Reset(watchRootParameterPollingInterval)
		}
	}
}
