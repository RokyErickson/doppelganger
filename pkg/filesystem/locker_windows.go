package filesystem

import (
	"syscall"
	"unsafe"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	lockFileEx   = kernel32.NewProc("LockFileEx")
	unlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	LOCKFILE_EXCLUSIVE_LOCK   = 2
	LOCKFILE_FAIL_IMMEDIATELY = 1
)

func callLockFileEx(
	handle syscall.Handle,
	flags,
	reserved,
	lockLow,
	lockHigh uint32,
	overlapped *syscall.Overlapped,
) (err error) {
	r1, _, e1 := syscall.Syscall6(
		lockFileEx.Addr(),
		6,
		uintptr(handle),
		uintptr(flags),
		uintptr(reserved),
		uintptr(lockLow),
		uintptr(lockHigh),
		uintptr(unsafe.Pointer(overlapped)),
	)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func callunlockFileEx(
	handle syscall.Handle,
	reserved,
	lockLow,
	lockHigh uint32,
	overlapped *syscall.Overlapped,
) (err error) {
	r1, _, e1 := syscall.Syscall6(
		unlockFileEx.Addr(),
		5,
		uintptr(handle),
		uintptr(reserved),
		uintptr(lockLow),
		uintptr(lockHigh),
		uintptr(unsafe.Pointer(overlapped)),
		0,
	)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func (l *Locker) Lock(block bool) error {
	var ol syscall.Overlapped
	flags := uint32(LOCKFILE_EXCLUSIVE_LOCK)
	if !block {
		flags |= LOCKFILE_FAIL_IMMEDIATELY
	}
	return callLockFileEx(syscall.Handle(l.file.Fd()), flags, 0, 1, 0, &ol)
}

func (l *Locker) Unlock() error {
	var ol syscall.Overlapped
	return callunlockFileEx(syscall.Handle(l.file.Fd()), 0, 1, 0, &ol)
}
