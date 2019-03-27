package sync

import (
	"github.com/pkg/errors"
)

func (a *Archive) EnsureValid() error {
	if a == nil {
		return errors.New("nil archive")
	}

	if err := a.Root.EnsureValid(); err != nil {
		return errors.Wrap(err, "invalid archive root")
	}

	return nil
}
