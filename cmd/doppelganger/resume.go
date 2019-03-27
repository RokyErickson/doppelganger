package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	promptpkg "github.com/RokyErickson/doppelganger/pkg/prompt"
	sessionsvcpkg "github.com/RokyErickson/doppelganger/pkg/service/session"
)

func resumeMain(command *cobra.Command, arguments []string) error {

	var specifications []string
	if len(arguments) > 0 {
		if resumeConfiguration.all {
			return errors.New("-a/--all specified with specific sessions")
		}
		specifications = arguments
	} else if !resumeConfiguration.all {
		return errors.New("no sessions specified")
	}

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	sessionService := sessionsvcpkg.NewSessionsClient(daemonConnection)

	resumeContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := sessionService.Resume(resumeContext)
	if err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to invoke resume")
	}

	request := &sessionsvcpkg.ResumeRequest{
		Specifications: specifications,
	}
	if err := stream.Send(request); err != nil {
		return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send resume request")
	}

	statusLinePrinter := &cmd.StatusLinePrinter{}

	for {
		if response, err := stream.Recv(); err != nil {
			statusLinePrinter.BreakIfNonEmpty()
			return errors.Wrap(peelAwayRPCErrorLayer(err), "resume failed")
		} else if err = response.EnsureValid(); err != nil {
			statusLinePrinter.BreakIfNonEmpty()
			return errors.Wrap(err, "invalid resume response received")
		} else if response.Message == "" && response.Prompt == "" {
			statusLinePrinter.Clear()
			return nil
		} else if response.Message != "" {
			statusLinePrinter.Print(response.Message)
			if err := stream.Send(&sessionsvcpkg.ResumeRequest{}); err != nil {
				statusLinePrinter.BreakIfNonEmpty()
				return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send message response")
			}
		} else if response.Prompt != "" {
			statusLinePrinter.BreakIfNonEmpty()
			if response, err := promptpkg.PromptCommandLine(response.Prompt); err != nil {
				return errors.Wrap(err, "unable to perform prompting")
			} else if err = stream.Send(&sessionsvcpkg.ResumeRequest{Response: response}); err != nil {
				return errors.Wrap(peelAwayRPCErrorLayer(err), "unable to send prompt response")
			}
		}
	}
}

var resumeCommand = &cobra.Command{
	Use:   "resume [<session>...]",
	Short: "Resumes a paused or disconnected synchronization session",
	Run:   cmd.Mainify(resumeMain),
}

var resumeConfiguration struct {
	help bool
	all  bool
}

func init() {

	flags := resumeCommand.Flags()
	flags.BoolVarP(&resumeConfiguration.help, "help", "h", false, "Show help information")
	flags.BoolVarP(&resumeConfiguration.all, "all", "a", false, "Resume all sessions")
}
