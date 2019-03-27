package process

import (
	"os/exec"
	"syscall"
)

const (
	DETACHED_PROCESS = 0x00000008
)

func DetachedProcessAttributes(process *exec.Cmd) {
	process.SysProcAttr = detachedprocessattributes()
}
func detachedprocessattributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: DETACHED_PROCESS,
	}
}
