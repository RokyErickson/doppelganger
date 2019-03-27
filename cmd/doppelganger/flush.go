package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	sessionsvcpkg "github.com/RokyErickson/doppelganger/pkg/service/session"
)

func flushMain(command *cobra.Command, arguments []string) error {
	var specifications []string
	if len(arguments) > 0 {
		if flushConfiguration.all {
			return errors.New("-a/--all specified with specific sessions")
		}
		specifications = arguments
	} else if !flushConfiguration.all {
		return errors.New("no sessions specified")
	}

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	sessionService := sessionsvcpkg.NewSessionsClient(daemonConnection)

	flushContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := sessionService.Flush(flushContext)
	if err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to invoke flush")
	}

	request := &sessionsvcpkg.FlushRequest{
		Specifications: specifications,
		SkipWait:       flushConfiguration.skipWait,
	}
	if err := stream.Send(request); err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send flush request")
	}

	statusLinePrinter := &cmd.StatusLinePrinter{}

	for {
		if response, err := stream.Recv(); err != nil {
			statusLinePrinter.BreakIfNonEmpty()
			return errors.Wrap(peelAwayRPCErrorLayer(err), "flush failed")
		} else if err = response.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid flush response received")
		} else if response.Message == "" {
			statusLinePrinter.Clear()
			return nil
		} else if response.Message != "" {
			statusLinePrinter.Print(response.Message)
			if err := stream.Send(&sessionsvcpkg.FlushRequest{}); err != nil {
				statusLinePrinter.BreakIfNonEmpty()
				return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send message response")
			}
		}
	}
}

var flushCommand = &cobra.Command{
	Use:   "flush [<session>...]",
	Short: "Flushes a synchronization session",
	Run:   cmd.Mainify(flushMain),
}

var flushConfiguration struct {
	help     bool
	all      bool
	skipWait bool
}

func init() {
	flags := flushCommand.Flags()
	flags.BoolVarP(&flushConfiguration.help, "help", "h", false, "Show help information")
	flags.BoolVarP(&flushConfiguration.all, "all", "a", false, "Flush all sessions")
	flags.BoolVar(&flushConfiguration.skipWait, "skip-wait", false, "Avoid waiting for the resulting synchronization cycle to complete")
}
