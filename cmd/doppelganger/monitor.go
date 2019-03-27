package main

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/fatih/color"

	"github.com/RokyErickson/doppelganger/cmd"
	sessionsvcpkg "github.com/RokyErickson/doppelganger/pkg/service/session"
	sessionpkg "github.com/RokyErickson/doppelganger/pkg/session"
)

func computeMonitorStatusLine(state *sessionpkg.State) string {

	status := "Status: "
	if state.Session.Paused {
		status += color.YellowString("[Paused]")
	} else {

		if len(state.Conflicts) > 0 {
			status += color.RedString("[Conflicts] ")
		}

		if len(state.AlphaProblems) > 0 || len(state.BetaProblems) > 0 {
			status += color.RedString("[Problems] ")
		}

		if state.LastError != "" {
			status += color.RedString("[Errored] ")
		}

		status += state.Status.Description()

		if (state.Status == sessionpkg.Status_StagingAlpha ||
			state.Status == sessionpkg.Status_StagingBeta) &&
			state.StagingStatus != nil {
			status += fmt.Sprintf(
				": %.0f%% (%d/%d)",
				100.0*float32(state.StagingStatus.Received)/float32(state.StagingStatus.Total),
				state.StagingStatus.Received,
				state.StagingStatus.Total,
			)
		}
	}

	return status
}

func monitorMain(command *cobra.Command, arguments []string) error {
	var session string
	var specifications []string
	if len(arguments) == 1 {
		session = arguments[0]
		specifications = []string{session}
	} else if len(arguments) > 1 {
		return errors.New("multiple session specification not allowed")
	}

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	sessionService := sessionsvcpkg.NewSessionsClient(daemonConnection)

	statusLinePrinter := &cmd.StatusLinePrinter{}
	defer statusLinePrinter.BreakIfNonEmpty()

	var previousStateIndex uint64
	sessionInformationPrinted := false
	for {
		request := &sessionsvcpkg.ListRequest{
			PreviousStateIndex: previousStateIndex,
			Specifications:     specifications,
		}

		response, err := sessionService.List(context.Background(), request)
		if err != nil {
			return errors.Wrap(peelAwayRPCErrorLayer(err), "list failed")
		} else if err = response.EnsureValid(); err != nil {
			return errors.Wrap(err, "invalid list response received")
		}

		for _, s := range response.SessionStates {
			if err = s.EnsureValid(); err != nil {
				return errors.Wrap(err, "invalid session state detected in response")
			}
		}

		var state *sessionpkg.State
		previousStateIndex = response.StateIndex
		if session == "" {
			if len(response.SessionStates) == 0 {
				err = errors.New("no sessions exist")
			} else {
				state = response.SessionStates[len(response.SessionStates)-1]
				session = state.Session.Identifier
				specifications = []string{session}
			}
		} else if len(response.SessionStates) != 1 {
			err = errors.New("invalid list response")
		} else {
			state = response.SessionStates[0]
		}
		if err != nil {
			return err
		}

		if !sessionInformationPrinted {
			printSession(state, monitorConfiguration.long)
			if !monitorConfiguration.long {
				fmt.Println("Alpha:", state.Session.Alpha.Format("\n\t"))
				fmt.Println("Beta:", state.Session.Beta.Format("\n\t"))
			}

			sessionInformationPrinted = true
		}

		statusLine := computeMonitorStatusLine(state)

		statusLinePrinter.Print(statusLine)
	}
}

var monitorCommand = &cobra.Command{
	Use:   "monitor [<session>]",
	Short: "Shows a dynamic status display for the specified session",
	Run:   cmd.Mainify(monitorMain),
}

var monitorConfiguration struct {
	help bool
	long bool
}

func init() {
	flags := monitorCommand.Flags()
	flags.BoolVarP(&monitorConfiguration.help, "help", "h", false, "Show help information")
	flags.BoolVarP(&monitorConfiguration.long, "long", "l", false, "Show detailed session information")
}
