package sync

import (
	"github.com/pkg/errors"

	fs "github.com/RokyErickson/doppelganger/pkg/filesystem"
)

const (
	allReadWritePermissionMask = fs.ModePermissionUserRead | fs.ModePermissionUserWrite |
		fs.ModePermissionGroupRead | fs.ModePermissionGroupWrite |
		fs.ModePermissionOthersRead | fs.ModePermissionOthersWrite

	allExecutePermissionMask = fs.ModePermissionUserExecute |
		fs.ModePermissionGroupExecute |
		fs.ModePermissionOthersExecute
)

func EnsureDefaultFileModeValid(mode fs.Mode) error {
	if (mode & allReadWritePermissionMask) != mode {
		return errors.New("executability bits detected in file mode")
	}

	if mode == 0 {
		return errors.New("zero-value file permission mode specified")
	}

	return nil
}

func EnsureDefaultDirectoryModeValid(mode fs.Mode) error {
	if (mode & fs.ModePermissionsMask) != mode {
		return errors.New("non-permission bits detected in directory mode")
	}

	if mode == 0 {
		return errors.New("zero-value directory permission mode specified")
	}

	return nil
}

func anyExecutableBitSet(mode fs.Mode) bool {
	return (mode & allExecutePermissionMask) != 0
}

func markExecutableForReaders(mode fs.Mode) fs.Mode {
	if (mode & fs.ModePermissionUserRead) != 0 {
		mode |= fs.ModePermissionUserExecute
	}

	if (mode & fs.ModePermissionGroupRead) != 0 {
		mode |= fs.ModePermissionGroupExecute
	}

	if (mode & fs.ModePermissionOthersRead) != 0 {
		mode |= fs.ModePermissionOthersExecute
	}

	return mode
}
