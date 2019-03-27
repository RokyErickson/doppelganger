// +build !windows

package filesystem

import (
	"golang.org/x/sys/unix"
)

type Mode uint32

const (
	ModeTypeMask = Mode(unix.S_IFMT)

	ModeTypeDirectory = Mode(unix.S_IFDIR)

	ModeTypeFile = Mode(unix.S_IFREG)

	ModeTypeSymbolicLink = Mode(unix.S_IFLNK)
)
