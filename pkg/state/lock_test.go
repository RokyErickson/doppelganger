package state

import (
	"testing"
	"time"
)

func TestTrackingLock(t *testing.T) {

	tracker := NewTracker()

	lock := NewTrackingLock(tracker)

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

	lock.Lock()
	lock.Unlock()
	select {
	case value := <-handoff:
		if !value {
			t.Fatal("received failure on state tracking")
		}
	case <-time.After(trackerTestTimeout):
		t.Fatal("timeout failure on state tracking")
	}

	time.Sleep(trackerTestSleep)

	lock.Lock()
	lock.UnlockWithoutNotify()

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
