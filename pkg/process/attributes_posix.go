// +build !windows,!plan9

// TODO: Figure out what to do for Plan 9. It doesn't support Setsid.

package process

import (
	"os/exec"
	"syscall"
)

func DetachedProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = detachedprocessattributes()
}
func detachedprocessattributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
