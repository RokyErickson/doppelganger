package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/prompt"
	_ "github.com/RokyErickson/doppelganger/pkg/protocols/docker"
	_ "github.com/RokyErickson/doppelganger/pkg/protocols/ipfs"
	_ "github.com/RokyErickson/doppelganger/pkg/protocols/local"
	_ "github.com/RokyErickson/doppelganger/pkg/protocols/ssh"
)

func rootMain(command *cobra.Command, arguments []string) error {
	command.Help()

	return nil
}

var rootCommand = &cobra.Command{
	Use:   "doppelganger",
	Short: "Doppelganger is a continous, straightforward, and bi-directional file synchronizer written in Go.",
	Run:   cmd.Mainify(rootMain),
}

var rootConfiguration struct {
	help bool
}

func init() {
	flags := rootCommand.Flags()

	flags.BoolVarP(&rootConfiguration.help, "help", "h", false, "Show help information")

	cobra.EnableCommandSorting = false

	cobra.MousetrapHelpText = ""

	rootCommand.AddCommand(
		createCommand,
		listCommand,
		monitorCommand,
		flushCommand,
		pauseCommand,
		resumeCommand,
		terminateCommand,
		daemonCommand,
		versionCommand,
		legalCommand,
		generateCommand,
	)
}

func main() {

	if _, ok := os.LookupEnv(prompt.PrompterEnvironmentVariable); ok {
		if err := promptMain(os.Args[1:]); err != nil {
			cmd.Fatal(err)
		}
		return
	}
	cmd.HandleTerminalCompatibility()

	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
