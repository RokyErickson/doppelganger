package daemon

import (
	"github.com/pkg/errors"

	"github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	lockName = "daemon.lock"
)

type Lock struct {
	locker *filesystem.Locker
}

func AcquireLock() (*Lock, error) {
	lockPath, err := subpath(lockName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compute daemon lock path")
	}

	locker, err := filesystem.NewLocker(lockPath, 0600)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create daemon locker")
	} else if err = locker.Lock(false); err != nil {
		return nil, err
	}

	return &Lock{
		locker: locker,
	}, nil
}

func (l *Lock) Unlock() error {
	return l.locker.Unlock()
}
