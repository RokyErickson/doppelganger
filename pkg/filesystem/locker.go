package filesystem

import (
	"os"

	"github.com/pkg/errors"
)

type Locker struct {
	file *os.File
}

func NewLocker(path string, permissions os.FileMode) (*Locker, error) {
	mode := os.O_RDWR | os.O_CREATE | os.O_APPEND
	if file, err := os.OpenFile(path, mode, permissions); err != nil {
		return nil, errors.Wrap(err, "unable to open lock file")
	} else {
		return &Locker{file: file}, nil
	}
}
