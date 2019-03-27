package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/agent"
	"github.com/RokyErickson/doppelganger/pkg/protocols/local"
	"github.com/RokyErickson/doppelganger/pkg/remote"
)

const (
	housekeepingInterval = 24 * time.Hour
)

func housekeep() {

	agent.Housekeep()

	local.HousekeepCaches()

	local.HousekeepStaging()
}

func housekeepRegularly(context context.Context) {

	housekeep()

	ticker := time.NewTicker(housekeepingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-context.Done():
			return
		case <-ticker.C:
			housekeep()
		}
	}
}

func endpointMain(command *cobra.Command, arguments []string) error {

	signalTermination := make(chan os.Signal, 1)
	signal.Notify(signalTermination, cmd.TerminationSignals...)

	housekeepingContext, housekeepingCancel := context.WithCancel(context.Background())
	defer housekeepingCancel()
	go housekeepRegularly(housekeepingContext)

	connection := newStdioConnection()

	endpointTermination := make(chan error, 1)
	go func() {
		endpointTermination <- remote.ServeEndpoint(connection)
	}()

	select {
	case sig := <-signalTermination:
		return errors.Errorf("terminated by signal: %s", sig)
	case err := <-endpointTermination:
		return errors.Wrap(err, "endpoint terminated")
	}
}

var endpointCommand = &cobra.Command{
	Use:   agent.ModeEndpoint,
	Short: "Run the agent in endpoint mode",
	Run:   cmd.Mainify(endpointMain),
}

var endpointConfiguration struct {
	help bool
}

func init() {

	flags := endpointCommand.Flags()
	flags.BoolVarP(&endpointConfiguration.help, "help", "h", false, "Show help information")
}
