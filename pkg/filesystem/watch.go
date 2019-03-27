package filesystem

import (
	"context"

	"github.com/pkg/errors"
)

func (m WatchMode) IsDefault() bool {
	return m == WatchMode_WatchModeDefault
}

func (m *WatchMode) UnmarshalText(textBytes []byte) error {

	text := string(textBytes)

	switch text {
	case "portable":
		*m = WatchMode_WatchModePortable
	case "force-poll":
		*m = WatchMode_WatchModeForcePoll
	case "no-watch":
		*m = WatchMode_WatchModeNoWatch
	default:
		return errors.Errorf("unknown watch mode specification: %s", text)
	}

	return nil
}

func (m WatchMode) Supported() bool {
	switch m {
	case WatchMode_WatchModePortable:
		return true
	case WatchMode_WatchModeForcePoll:
		return true
	case WatchMode_WatchModeNoWatch:
		return true
	default:
		return false
	}
}

func (m WatchMode) Description() string {
	switch m {
	case WatchMode_WatchModeDefault:
		return "Default"
	case WatchMode_WatchModePortable:
		return "Portable"
	case WatchMode_WatchModeForcePoll:
		return "Force Poll"
	case WatchMode_WatchModeNoWatch:
		return "No Watch"
	default:
		return "Unknown"
	}
}

func Watch(context context.Context, root string, events chan struct{}, mode WatchMode, pollInterval uint32) {

	if cap(events) < 1 {
		panic("watch channel should be buffered")
	} else if pollInterval == 0 {
		panic("polling interval must be greater than 0 seconds")
	}

	defer close(events)

	if mode == WatchMode_WatchModeNoWatch {
		<-context.Done()
		return
	}

	if mode == WatchMode_WatchModePortable {
		watchNative(context, root, events, pollInterval)
	}

	select {
	case <-context.Done():
		return
	default:
	}

	watchPoll(context, root, events, pollInterval)
}
