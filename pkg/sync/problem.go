package sync

import (
	"github.com/pkg/errors"
)

func (p *Problem) EnsureValid() error {
	if p == nil {
		return errors.New("nil problem")
	}

	return nil
}
