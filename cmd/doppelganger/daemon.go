package main

import (
	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/daemon"
)

func daemonMain(command *cobra.Command, arguments []string) error {

	command.Help()
	return nil
}

var daemonCommand = &cobra.Command{
	Use:   "daemon",
	Short: "Controls the Doppelganger daemon lifecycle",
	Run:   cmd.Mainify(daemonMain),
}

var daemonConfiguration struct {
	help bool
}

func init() {

	flags := daemonCommand.Flags()
	flags.BoolVarP(&daemonConfiguration.help, "help", "h", false, "Show help information")
	if daemon.RegistrationSupported {
		daemonCommand.AddCommand(
			daemonRunCommand,
			daemonStartCommand,
			daemonStopCommand,
			daemonRegisterCommand,
			daemonUnregisterCommand,
		)
	} else {
		daemonCommand.AddCommand(
			daemonRunCommand,
			daemonStartCommand,
			daemonStopCommand,
		)
	}
}
