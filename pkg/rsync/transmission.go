package rsync

import (
	"github.com/pkg/errors"
)

func (t *Transmission) resetToZeroMaintainingCapacity() {

	t.Done = false

	if t.Operation != nil {
		t.Operation.resetToZeroMaintainingCapacity()
	}

	t.Error = ""
}

func (t *Transmission) EnsureValid() error {

	if t == nil {
		return errors.New("nil transmission")
	}

	if t.Done {
		if t.Operation != nil && !t.Operation.isZeroValue() {
			return errors.New("operation present at end of stream")
		}
	} else {
		if t.Operation == nil {
			return errors.New("operation missing from middle of stream")
		} else if err := t.Operation.EnsureValid(); err != nil {
			return errors.New("invalid operation in stream")
		} else if t.Error != "" {
			return errors.New("error in middle of stream")
		}
	}

	return nil
}
