package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	sessionsvcpkg "github.com/RokyErickson/doppelganger/pkg/service/session"
)

func pauseMain(command *cobra.Command, arguments []string) error {

	var specifications []string
	if len(arguments) > 0 {
		if pauseConfiguration.all {
			return errors.New("-a/--all specified with specific sessions")
		}
		specifications = arguments
	} else if !pauseConfiguration.all {
		return errors.New("no sessions specified")
	}

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	sessionService := sessionsvcpkg.NewSessionsClient(daemonConnection)

	pauseContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := sessionService.Pause(pauseContext)
	if err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to invoke pause")
	}

	request := &sessionsvcpkg.PauseRequest{
		Specifications: specifications,
	}
	if err := stream.Send(request); err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send pause request")
	}

	statusLinePrinter := &cmd.StatusLinePrinter{}

	for {
		if response, err := stream.Recv(); err != nil {
			statusLinePrinter.BreakIfNonEmpty()
			return errors.Wrap(peelAwayRPCErrorLayer(err), "pause failed")
		} else if err = response.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid pause response received")
		} else if response.Message == "" {
			statusLinePrinter.Clear()
			return nil
		} else if response.Message != "" {
			statusLinePrinter.Print(response.Message)
			if err := stream.Send(&sessionsvcpkg.PauseRequest{}); err != nil {
				statusLinePrinter.BreakIfNonEmpty()
				return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send message response")
			}
		}
	}
}

var pauseCommand = &cobra.Command{
	Use:   "pause [<session>...]",
	Short: "Pauses a synchronization session",
	Run:   cmd.Mainify(pauseMain),
}

var pauseConfiguration struct {
	help bool
	all  bool
}

func init() {

	flags := pauseCommand.Flags()
	flags.BoolVarP(&pauseConfiguration.help, "help", "h", false, "Show help information")
	flags.BoolVarP(&pauseConfiguration.all, "all", "a", false, "Pause all sessions")
}
