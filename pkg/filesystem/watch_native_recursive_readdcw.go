// +build windows

package filesystem

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem/winfsnotify"
)

const (
	winfsnotifyFlags = winfsnotify.FS_ALL_EVENTS & ^(winfsnotify.FS_ACCESS | winfsnotify.FS_CLOSE)
)

type recursiveWatch struct {
	watcher *winfsnotify.Watcher

	forwardingCancel context.CancelFunc

	eventPaths chan string
}

func newRecursiveWatch(path string, _ os.FileInfo) (*recursiveWatch, error) {

	watcher, err := winfsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create watcher")
	}

	eventPaths := make(chan string, watchNativeEventsBufferSize)

	forwardingContext, forwardingCancel := context.WithCancel(context.Background())
	go func() {
	Forwarding:
		for {
			select {
			case <-forwardingContext.Done():
				break Forwarding
			case e, ok := <-watcher.Event:
				if !ok || e.Mask == winfsnotify.FS_Q_OVERFLOW {
					break Forwarding
				}
				select {
				case eventPaths <- e.Name:
				default:
				}
			}
		}
		close(eventPaths)
	}()

	if err := watcher.AddWatch(path, winfsnotifyFlags); err != nil {
		forwardingCancel()
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, errors.Wrap(err, "unable to start watching")
	}

	return &recursiveWatch{
		watcher:          watcher,
		forwardingCancel: forwardingCancel,
		eventPaths:       eventPaths,
	}, nil
}

func (w *recursiveWatch) stop() {

	w.watcher.Close()

	w.forwardingCancel()
}
