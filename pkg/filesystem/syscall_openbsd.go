package filesystem

import (
	"golang.org/x/sys/unix"
)

const (
	_AT_REMOVEDIR = 0x8
)

func mkdirat(directory int, path string, mode uint32) error {
	return unix.Mkdirat(directory, path, mode)
}

func symlinkat(target string, directory int, path string) error {
	return unix.Symlinkat(target, directory, path)
}

func readlinkat(directory int, path string, buffer []byte) (int, error) {
	return unix.Readlinkat(directory, path, buffer)
}

func openat(directory int, path string, flags int, mode uint32) (int, error) {
	return unix.Openat(directory, path, flags, mode)
}

func lstat(path string, metadata *unix.Stat_t) error {
	return unix.Lstat(path, metadata)
}

func fstatat(directory int, path string, metadata *unix.Stat_t, flags int) error {
	return unix.Fstatat(directory, path, metadata, flags)
}

func fchmodat(directory int, path string, mode uint32, flags int) error {
	return unix.Fchmodat(directory, path, mode, flags)
}

func fchownat(directory int, path string, userId, groupId, flags int) error {
	return unix.Fchownat(directory, path, userId, groupId, flags)
}

func renameat(sourceDirectory int, sourcePath string, targetDirectory int, targetPath string) error {
	return unix.Renameat(sourceDirectory, sourcePath, targetDirectory, targetPath)
}

func unlinkat(directory int, path string, flags int) error {
	return unix.Unlinkat(directory, path, flags)
}
