package filesystem

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	_AT_REMOVEDIR = unix.AT_REMOVEDIR
)

type libcFunction uintptr

func sysvicall6(trap, nargs, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2 uintptr, err syscall.Errno)

var (
	procSymlinkat,

	procReadlinkat libcFunction
)

func mkdirat(directory int, path string, mode uint32) error {
	return unix.Mkdirat(directory, path, mode)
}

func symlinkat(target string, directory int, path string) error {

	var targetBytes *byte
	targetBytes, err := unix.BytePtrFromString(target)
	if err != nil {
		return err
	}

	var pathBytes *byte
	pathBytes, err = unix.BytePtrFromString(path)
	if err != nil {
		return err
	}

	_, _, errnoErr := sysvicall6(uintptr(unsafe.Pointer(&procSymlinkat)), 3, uintptr(unsafe.Pointer(targetBytes)), uintptr(directory), uintptr(unsafe.Pointer(pathBytes)), 0, 0, 0)
	if errnoErr != 0 {
		return errnoErr
	}

	return nil
}

func readlinkat(directory int, path string, buffer []byte) (int, error) {

	var pathBytes *byte
	pathBytes, err := unix.BytePtrFromString(path)
	if err != nil {
		return 0, err
	}

	var bufferBytes *byte
	if len(buffer) > 0 {
		bufferBytes = &buffer[0]
	}

	n, _, errnoErr := sysvicall6(uintptr(unsafe.Pointer(&procReadlinkat)), 4, uintptr(directory), uintptr(unsafe.Pointer(pathBytes)), uintptr(unsafe.Pointer(bufferBytes)), uintptr(len(buffer)), 0, 0)
	if errnoErr != 0 {
		return 0, errnoErr
	}

	return int(n), nil
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
