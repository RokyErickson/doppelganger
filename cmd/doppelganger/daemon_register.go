package main

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/daemon"
)

func daemonRegisterMain(command *cobra.Command, arguments []string) error {
	if len(arguments) != 0 {
		return errors.New("unexpected arguments provided")
	}

	if err := daemon.Register(); err != nil {
		return err
	}

	return nil
}

var daemonRegisterCommand = &cobra.Command{
	Use:   "register",
	Short: "Registers Doppelganger to start as a per-user daemon on login",
	Run:   cmd.Mainify(daemonRegisterMain),
}

var daemonRegisterConfiguration struct {
	help bool
}

func init() {

	flags := daemonRegisterCommand.Flags()
	flags.BoolVarP(&daemonRegisterConfiguration.help, "help", "h", false, "Show help information")
}
