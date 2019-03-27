// +build linux

package filesystem

import (
	"context"

	"github.com/RokyErickson/doppelganger/pkg/filesystem/notify"
)

type nonRecursiveWatcher struct {
	watcher notify.Watcher

	forwardingCancel context.CancelFunc

	eventPaths chan string
}

func newNonRecursiveWatcher() (*nonRecursiveWatcher, error) {

	rawEvents := make(chan notify.EventInfo, watchNativeEventsBufferSize)

	eventPaths := make(chan string, watchNativeEventsBufferSize)

	forwardingContext, forwardingCancel := context.WithCancel(context.Background())
	go func() {
	Forwarding:
		for {
			select {
			case <-forwardingContext.Done():
				break Forwarding
			case e, ok := <-rawEvents:
				if !ok {
					break Forwarding
				}
				select {
				case eventPaths <- e.Path():
				default:
				}
			}
		}
		close(eventPaths)
	}()

	watcher := notify.NewWatcher(rawEvents)

	return &nonRecursiveWatcher{
		watcher:          watcher,
		forwardingCancel: forwardingCancel,
		eventPaths:       eventPaths,
	}, nil
}

func (w *nonRecursiveWatcher) watch(path string) error {
	return w.watcher.Watch(
		path,
		notify.InModify|notify.InAttrib|
			notify.InCloseWrite|
			notify.InMovedFrom|notify.InMovedTo|
			notify.InCreate|notify.InDelete|
			notify.InDeleteSelf|notify.InMoveSelf,
	)
}

func (w *nonRecursiveWatcher) unwatch(path string) error {
	return w.watcher.Unwatch(path)
}

func (w *nonRecursiveWatcher) stop() {

	w.watcher.Close()

	w.forwardingCancel()
}
