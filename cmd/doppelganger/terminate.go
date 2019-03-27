package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	sessionsvcpkg "github.com/RokyErickson/doppelganger/pkg/service/session"
)

func terminateMain(command *cobra.Command, arguments []string) error {

	var specifications []string
	if len(arguments) > 0 {
		if terminateConfiguration.all {
			return errors.New("-a/--all specified with specific sessions")
		}
		specifications = arguments
	} else if !terminateConfiguration.all {
		return errors.New("no sessions specified")
	}

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	sessionService := sessionsvcpkg.NewSessionsClient(daemonConnection)

	terminateContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := sessionService.Terminate(terminateContext)
	if err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to invoke terminate")
	}

	request := &sessionsvcpkg.TerminateRequest{
		Specifications: specifications,
	}
	if err := stream.Send(request); err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send terminate request")
	}

	statusLinePrinter := &cmd.StatusLinePrinter{}

	for {
		if response, err := stream.Recv(); err != nil {
			statusLinePrinter.BreakIfNonEmpty()
			return errors.Wrap(peelAwayRPCErrorLayer(err), "terminate failed")
		} else if err = response.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid terminate response received")
		} else if response.Message == "" {
			statusLinePrinter.Clear()
			return nil
		} else if response.Message != "" {
			statusLinePrinter.Print(response.Message)
			if err := stream.Send(&sessionsvcpkg.TerminateRequest{}); err != nil {
				statusLinePrinter.BreakIfNonEmpty()
				return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send message response")
			}
		}
	}
}

var terminateCommand = &cobra.Command{
	Use:   "terminate [<session>...]",
	Short: "Permanently terminates a synchronization session",
	Run:   cmd.Mainify(terminateMain),
}

var terminateConfiguration struct {
	help bool
	all  bool
}

func init() {
	flags := terminateCommand.Flags()
	flags.BoolVarP(&terminateConfiguration.help, "help", "h", false, "Show help information")
	flags.BoolVarP(&terminateConfiguration.all, "all", "a", false, "Terminate all sessions")
}
