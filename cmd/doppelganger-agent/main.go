package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
)

func rootMain(command *cobra.Command, arguments []string) error {

	command.Help()

	return nil
}

var rootCommand = &cobra.Command{
	Use:   "doppelganger-agent",
	Short: "The Doppelganger agent should not be invoked by mere mortalss.",
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
		installCommand,
		endpointCommand,
		versionCommand,
		legalCommand,
	)
}

func main() {

	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
