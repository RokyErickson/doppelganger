// +build !plan9

// TODO: Figure out what to do for Plan 9. It doesn't have syscall.WaitStatus.

package process

import (
	"os"
	"syscall"

	"github.com/pkg/errors"
)

const (
	posixShellInvalidCommandExitCode = 126

	posixShellCommandNotFoundExitCode = 127
)

func ExitCodeForProcessState(state *os.ProcessState) (int, error) {

	waitStatus, ok := state.Sys().(syscall.WaitStatus)
	if !ok {
		return 0, errors.New("unable to access wait status")
	}

	return waitStatus.ExitStatus(), nil
}

func IsPOSIXShellInvalidCommand(state *os.ProcessState) bool {

	code, err := ExitCodeForProcessState(state)

	return err == nil && code == posixShellInvalidCommandExitCode
}

func IsPOSIXShellCommandNotFound(state *os.ProcessState) bool {

	code, err := ExitCodeForProcessState(state)

	return err == nil && code == posixShellCommandNotFoundExitCode
}
