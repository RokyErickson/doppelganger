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
	"github.com/RokyErickson/doppelganger/pkg/sync"
	urlpkg "github.com/RokyErickson/doppelganger/pkg/url"
)

func formatPath(path string) string {
	if path == "" {
		return "<root>"
	}
	return path
}

func formatConnectionStatus(connected bool) string {
	if connected {
		return "Connected"
	}
	return "Disconnected"
}

func printEndpointStatus(name string, url *urlpkg.URL, connected bool, problems []*sync.Problem) {
	fmt.Printf("%s:\n", name)
	if !listConfiguration.long {
		fmt.Println("\tURL:", url.Format("\n\t\t"))
	}

	fmt.Printf("\tConnection state: %s\n", formatConnectionStatus(connected))

	if len(problems) > 0 {
		color.Red("\tProblems:\n")
		for _, p := range problems {
			color.Red("\t\t%s: %v\n", formatPath(p.Path), p.Error)
		}
	}
}

func printSessionStatus(state *sessionpkg.State) {
	statusString := state.Status.Description()
	if state.Session.Paused {
		statusString = color.YellowString("[Paused]")
	}
	fmt.Fprintln(color.Output, "Status:", statusString)

	if state.LastError != "" {
		color.Red("Last error: %s\n", state.LastError)
	}
}

func formatEntryKind(entry *sync.Entry) string {
	if entry == nil {
		return "<non-existent>"
	} else if entry.Kind == sync.EntryKind_Directory {
		return "Directory"
	} else if entry.Kind == sync.EntryKind_File {
		if entry.Executable {
			return fmt.Sprintf("Executable File (%x)", entry.Digest)
		}
		return fmt.Sprintf("File (%x)", entry.Digest)
	} else if entry.Kind == sync.EntryKind_Symlink {
		return fmt.Sprintf("Symbolic Link (%s)", entry.Target)
	} else {
		return "<unknown>"
	}
}

func printConflicts(conflicts []*sync.Conflict) {

	color.Red("Conflicts:\n")

	for i, c := range conflicts {

		for _, a := range c.AlphaChanges {
			color.Red(
				"\t(α) %s (%s -> %s)\n",
				formatPath(a.Path),
				formatEntryKind(a.Old),
				formatEntryKind(a.New),
			)
		}

		for _, b := range c.BetaChanges {
			color.Red(
				"\t(β) %s (%s -> %s)\n",
				formatPath(b.Path),
				formatEntryKind(b.Old),
				formatEntryKind(b.New),
			)
		}

		if i < len(conflicts)-1 {
			fmt.Println()
		}
	}
}

func listMain(command *cobra.Command, arguments []string) error {

	daemonConnection, err := createDaemonClientConnection()
	if err != nil {
		return errors.Wrap(err, "unable to connect to daemon")
	}
	defer daemonConnection.Close()

	sessionService := sessionsvcpkg.NewSessionsClient(daemonConnection)

	request := &sessionsvcpkg.ListRequest{
		Specifications: arguments,
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

	for _, state := range response.SessionStates {
		fmt.Println(delimiterLine)
		printSession(state, listConfiguration.long)
		printEndpointStatus("Alpha", state.Session.Alpha, state.AlphaConnected, state.AlphaProblems)
		printEndpointStatus("Beta", state.Session.Beta, state.BetaConnected, state.BetaProblems)
		printSessionStatus(state)
		if len(state.Conflicts) > 0 {
			printConflicts(state.Conflicts)
		}
	}

	if len(response.SessionStates) > 0 {
		fmt.Println(delimiterLine)
	}

	return nil
}

var listCommand = &cobra.Command{
	Use:   "list [<session>...]",
	Short: "Lists existing synchronization sessions and their statuses",
	Run:   cmd.Mainify(listMain),
}

var listConfiguration struct {
	help bool
	long bool
}

func init() {

	flags := listCommand.Flags()
	flags.BoolVarP(&listConfiguration.help, "help", "h", false, "Show help information")
	flags.BoolVarP(&listConfiguration.long, "long", "l", false, "Show detailed session information")
}
