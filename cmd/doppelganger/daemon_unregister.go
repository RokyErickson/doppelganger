package main

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/daemon"
)

func daemonUnregisterMain(command *cobra.Command, arguments []string) error {

	if len(arguments) != 0 {
		return errors.New("unexpected arguments provided")
	}

	if err := daemon.Unregister(); err != nil {
		return err
	}

	return nil
}

var daemonUnregisterCommand = &cobra.Command{
	Use:   "unregister",
	Short: "Unregisters Doppelganger as a per-user daemon",
	Run:   cmd.Mainify(daemonUnregisterMain),
}

var daemonUnregisterConfiguration struct {
	help bool
}

func init() {
	flags := daemonUnregisterCommand.Flags()
	flags.BoolVarP(&daemonUnregisterConfiguration.help, "help", "h", false, "Show help information")
}
