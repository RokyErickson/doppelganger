package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/daemon"
	daemonsvcpkg "github.com/RokyErickson/doppelganger/pkg/service/daemon"
)

func daemonStopMain(command *cobra.Command, arguments []string) error {
	if len(arguments) != 0 {
		return errors.New("unexpected arguments provided")
	}

	if handled, err := daemon.RegisteredStop(); err != nil {
		return errors.Wrap(err, "unable to stop daemon using system mechanism")
	} else if handled {
		return nil
	}

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	daemonService := daemonsvcpkg.NewDaemonClient(daemonConnection)

	daemonService.Terminate(context.Background(), &daemonsvcpkg.TerminateRequest{})

	return nil
}

var daemonStopCommand = &cobra.Command{
	Use:   "stop",
	Short: "Stops the Doppelganger daemon if it's running",
	Run:   cmd.Mainify(daemonStopMain),
}

var daemonStopConfiguration struct {
	help bool
}

func init() {
	flags := daemonStopCommand.Flags()
	flags.BoolVarP(&daemonStopConfiguration.help, "help", "h", false, "Show help information")
}
