package cmd

import (
	"os"
	"syscall"
)

var TerminationSignals = []os.Signal{

	syscall.SIGINT,
}
