package cmd

import (
	"github.com/pkg/errors"
	"github.com/polydawn/gosh"
	"os"

	isatty "github.com/mattn/go-isatty"

	"github.com/RokyErickson/doppelganger/pkg/process"
)

func HandleTerminalCompatibility() {

	if !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return
	}

	executable, err := os.Executable()
	if err != nil {
		Fatal(errors.Wrap(err, "running inside mintty terminal and unable to locate current executable"))
	}

	arguments := make([]string, 0, len(os.Args))
	arguments = append(arguments, executable)
	arguments = append(arguments, os.Args[1:]...)

	command := gosh.Gosh("winpty", arguments)
	command.Run()
}
