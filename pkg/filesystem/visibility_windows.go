package filesystem

import (
	"syscall"

	"github.com/pkg/errors"
)

func markHidden(path string) error {

	path16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return errors.Wrap(err, "unable to convert path encoding")
	}

	attributes, err := syscall.GetFileAttributes(path16)
	if err != nil {
		return errors.Wrap(err, "unable to get file attributes")
	}

	attributes |= syscall.FILE_ATTRIBUTE_HIDDEN

	err = syscall.SetFileAttributes(path16, attributes)
	if err != nil {
		return errors.Wrap(err, "unable to set file attributes")
	}

	return nil
}
