package state

import (
	"sync"
)

type Tracker struct {
	change   *sync.Cond
	index    uint64
	poisoned bool
}

func NewTracker() *Tracker {
	return &Tracker{
		change: sync.NewCond(&sync.Mutex{}),
		index:  1,
	}
}

func (t *Tracker) Poison() {
	t.change.L.Lock()
	defer t.change.L.Unlock()

	t.poisoned = true
	t.change.Broadcast()
}

func (t *Tracker) NotifyOfChange() {

	t.change.L.Lock()
	defer t.change.L.Unlock()

	t.index += 1
	t.change.Broadcast()
}

func (t *Tracker) WaitForChange(previousIndex uint64) (uint64, bool) {

	t.change.L.Lock()
	defer t.change.L.Unlock()

	for t.index == previousIndex && !t.poisoned {
		t.change.Wait()
	}
	return t.index, t.poisoned
}
