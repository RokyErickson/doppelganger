package main

import (
	"github.com/polydawn/gosh"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/daemon"
)

func daemonStartMain(command *cobra.Command, arguments []string) error {
	if len(arguments) != 0 {
		return errors.New("unexpected arguments provided")
	}

	if handled, err := daemon.RegisteredStart(); err != nil {
		return errors.Wrap(err, "unable to start daemon using system mechanism")
	} else if handled {
		return nil
	}

	gosh.Gosh([]string{"doppelganger", "daemon", "run"}).Start()

	return nil
}

var daemonStartCommand = &cobra.Command{
	Use:   "start",
	Short: "Starts the Doppelganger daemon if it's not already running",
	Run:   cmd.Mainify(daemonStartMain),
}

var daemonStartConfiguration struct {
	help bool
}

func init() {
	flags := daemonStartCommand.Flags()
	flags.BoolVarP(&daemonStartConfiguration.help, "help", "h", false, "Show help information")
}
