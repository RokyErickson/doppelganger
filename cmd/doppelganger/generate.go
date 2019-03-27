package main

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
)

func generateMain(command *cobra.Command, arguments []string) error {
	if len(arguments) > 0 {
		return errors.New("this command does not accept arguments")
	}

	flagSpecified := generateConfiguration.bashCompletionScript != ""
	if !flagSpecified {
		return errors.New("no flags specified")
	}

	if generateConfiguration.bashCompletionScript != "" {
		if err := rootCommand.GenBashCompletionFile(generateConfiguration.bashCompletionScript); err != nil {
			return errors.Wrap(err, "unable to generate bash completion script")
		}
	}

	return nil
}

var generateCommand = &cobra.Command{
	Use:    "generate",
	Short:  "Generate various files",
	Run:    cmd.Mainify(generateMain),
	Hidden: true,
}

var generateConfiguration struct {
	help                 bool
	bashCompletionScript string
}

func init() {
	flags := generateCommand.Flags()
	flags.BoolVarP(&generateConfiguration.help, "help", "h", false, "Show help information")
	flags.StringVar(&generateConfiguration.bashCompletionScript, "bash-completion-script", "", "Generate bash completion script")
}
