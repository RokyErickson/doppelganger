// +build darwin,cgo

package filesystem

import (
	"context"
	"os"
	"time"

	"github.com/havoc-io/fsevents"
)

const (
	fseventsCoalescingLatency = 10 * time.Millisecond

	fseventsFlags = fsevents.WatchRoot | fsevents.FileEvents
)

type recursiveWatch struct {
	eventStream *fsevents.EventStream

	forwardingCancel context.CancelFunc

	eventPaths chan string
}

func newRecursiveWatch(path string, info os.FileInfo) (*recursiveWatch, error) {

	rawEvents := make(chan []fsevents.Event, watchNativeEventsBufferSize)

	eventStream := &fsevents.EventStream{
		Events:  rawEvents,
		Paths:   []string{path},
		Latency: fseventsCoalescingLatency,
		Flags:   fseventsFlags,
	}

	eventPaths := make(chan string, watchNativeEventsBufferSize)

	forwardingContext, forwardingCancel := context.WithCancel(context.Background())
	go func() {
	Forwarding:
		for {
			select {
			case <-forwardingContext.Done():
				break Forwarding
			case es, ok := <-rawEvents:
				if !ok {
					break Forwarding
				}
				for _, e := range es {
					select {
					case eventPaths <- e.Path:
					default:
					}
				}
			}
		}
		close(eventPaths)
	}()

	eventStream.Start()

	return &recursiveWatch{
		eventStream:      eventStream,
		forwardingCancel: forwardingCancel,
		eventPaths:       eventPaths,
	}, nil
}

func (w *recursiveWatch) stop() {

	w.eventStream.Stop()

	w.forwardingCancel()
}
