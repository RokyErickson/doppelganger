package sync

import (
	"github.com/pkg/errors"
)

func (c *Change) copySlim() *Change {
	return &Change{
		Path: c.Path,
		Old:  c.Old.copySlim(),
		New:  c.New.copySlim(),
	}
}

func (c *Change) EnsureValid() error {
	if c == nil {
		return errors.New("nil change")
	}

	return nil
}
