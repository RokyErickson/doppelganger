// +build linux

package filesystem

import (
	contextpkg "context"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/golang/groupcache/lru"
)

const (
	watchNativeNonRecursiveMaximumWatches = 50
)

func watchNative(context contextpkg.Context, root string, events chan struct{}, pollInterval uint32) error {

	if pollInterval == 0 {
		return errors.New("polling interval must be greater than 0 seconds")
	}
	pollIntervalDuration := time.Duration(pollInterval) * time.Second

	pollingTimer := time.NewTimer(0)

	coalescingTimer := time.NewTimer(watchNativeCoalescingWindow)
	if !coalescingTimer.Stop() {
		<-coalescingTimer.C
	}

	rootParentPath, rootLeafName := filepath.Split(root)

	rootParentWatcher, err := newNonRecursiveWatcher()
	if err != nil {
		return errors.Wrap(err, "unable to create root parent watcher")
	}
	defer rootParentWatcher.stop()

	var rootParentExists bool
	var rootParentMetadata os.FileInfo
	var rootParentWatched bool

	watcher, err := newNonRecursiveWatcher()
	if err != nil {
		return errors.Wrap(err, "unable to create watcher")
	}
	defer watcher.stop()

	watchedPaths := make(map[string]os.FileInfo, watchNativeNonRecursiveMaximumWatches-1)

	unwatchErrors := make(chan error, 1)
	watchedPathManager := lru.New(watchNativeNonRecursiveMaximumWatches - 1)
	watchedPathManager.OnEvicted = func(key lru.Key, _ interface{}) {
		if path, ok := key.(string); !ok {
			panic("invalid key type in watch path cache")
		} else {
			if err := watcher.unwatch(path); err != nil {
				select {
				case unwatchErrors <- err:
				default:
				}
			}
			delete(watchedPaths, path)
		}
	}

	monitoringContext, monitoringCancel := contextpkg.WithCancel(contextpkg.Background())
	defer monitoringCancel()
	monitoringErrors := make(chan error, 1)
	go func() {
		for {
			var resetCoalescingTimer bool
			select {
			case <-monitoringContext.Done():
				monitoringErrors <- errors.New("monitoring cancelled")
				return
			case path, ok := <-rootParentWatcher.eventPaths:
				if !ok {
					monitoringErrors <- errors.New("root parent watcher event stream closed")
					return
				} else if filepath.Base(path) == rootLeafName {
					resetCoalescingTimer = true
				}
			case path, ok := <-watcher.eventPaths:
				if !ok {
					monitoringErrors <- errors.New("watcher event stream closed")
					return
				}
				resetCoalescingTimer = !IsTemporaryFileName(filepath.Base(path))
			}

			if resetCoalescingTimer {
				if !coalescingTimer.Stop() {
					select {
					case <-coalescingTimer.C:
					default:
					}
				}
				coalescingTimer.Reset(watchNativeCoalescingWindow)
			}
		}
	}()

	var contents map[string]os.FileInfo

	for {
		select {
		case <-context.Done():
			return errors.New("watch cancelled")
		case err := <-unwatchErrors:
			return errors.Wrap(err, "unable to unwatch path")
		case <-coalescingTimer.C:
			select {
			case events <- struct{}{}:
			default:
			}
		case <-pollingTimer.C:
			newContents, changed, changes, err := poll(root, contents, true)
			if err != nil {
				pollingTimer.Reset(pollIntervalDuration)
				continue
			}

			contents = newContents

			if changed {
				select {
				case events <- struct{}{}:
				default:
				}
			}

			var rootParentCurrentlyExists bool
			var currentRootParentMetadata os.FileInfo
			if m, err := os.Lstat(rootParentPath); err != nil {
				if !os.IsNotExist(err) {
					return errors.Wrap(err, "unable to probe root parent metadata")
				}
			} else {
				rootParentCurrentlyExists = true
				currentRootParentMetadata = m
			}

			reestablishRootParentWatch := rootParentCurrentlyExists != rootParentExists ||
				!watchRootParametersEqual(currentRootParentMetadata, rootParentMetadata)

			if reestablishRootParentWatch {

				if rootParentWatched {
					if err := rootParentWatcher.unwatch(rootParentPath); err != nil {
						return errors.Wrap(err, "unable to remove stale root parent watch")
					}
					rootParentWatched = false
				}

				if rootParentCurrentlyExists {
					if err := rootParentWatcher.watch(rootParentPath); err != nil {
						if os.IsNotExist(err) {
							rootParentCurrentlyExists = false
							currentRootParentMetadata = nil
						} else {
							return errors.Wrap(err, "unable to watch root parent path")
						}
					} else {
						rootParentWatched = true
					}
				}

			}

			rootParentExists = rootParentCurrentlyExists
			rootParentMetadata = currentRootParentMetadata

			for p, m := range watchedPaths {
				currentMetadata, ok := newContents[p]
				if ok && watchRootParametersEqual(currentMetadata, m) {
					continue
				}
				watchedPathManager.Remove(p)
			}

			for p := range changes {
				if _, ok := watchedPaths[p]; ok {
					delete(changes, p)
				}
			}

			if len(changes) > (watchNativeNonRecursiveMaximumWatches - 1) {
				changes = nil
			}

			for p := range changes {
				if err := watcher.watch(p); err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return errors.Wrap(err, "unable to create watch")
				}
				watchedPaths[p] = newContents[p]
				watchedPathManager.Add(p, 0)
			}

			pollingTimer.Reset(pollIntervalDuration)
		}
	}
}
