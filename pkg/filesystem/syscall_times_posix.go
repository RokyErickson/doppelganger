// +build !windows,!darwin,!netbsd

package filesystem

import (
	"golang.org/x/sys/unix"
)

func extractModificationTime(metadata *unix.Stat_t) unix.Timespec {
	return metadata.Mtim
}
