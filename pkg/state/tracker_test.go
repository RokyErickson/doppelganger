package state

import (
	"testing"
	"time"
)

const (
	trackerTestSleep   = 10 * time.Millisecond
	trackerTestTimeout = 1 * time.Second
)

func TestTracker(t *testing.T) {
	tracker := NewTracker()
	handoff := make(chan bool)

	go func() {
		firstState, poisoned := tracker.WaitForChange(1)
		if poisoned || firstState != 2 {
			handoff <- false
			return
		}
		handoff <- true

		_, poisoned = tracker.WaitForChange(firstState)
		handoff <- poisoned
	}()

	tracker.NotifyOfChange()
	select {
	case value := <-handoff:
		if !value {
			t.Fatal("received failure on state tracking")
		}
	case <-time.After(trackerTestTimeout):
		t.Fatal("timeout failure on state tracking")
	}

	time.Sleep(trackerTestSleep)

	tracker.Poison()
	select {
	case value := <-handoff:
		if !value {
			t.Fatal("received failure on state poisoning")
		}
	case <-time.After(trackerTestTimeout):
		t.Fatal("timeout failure on state poisoning")
	}
}
