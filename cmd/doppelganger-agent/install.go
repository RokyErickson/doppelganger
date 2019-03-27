package main

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/agent"
)

func installMain(command *cobra.Command, arguments []string) error {
	return errors.Wrap(agent.Install(), "installation error")
}

var installCommand = &cobra.Command{
	Use:   agent.ModeInstall,
	Short: "Perform agent installation",
	Run:   cmd.Mainify(installMain),
}

var installConfiguration struct {
	help bool
}

func init() {
	flags := installCommand.Flags()
	flags.BoolVarP(&installConfiguration.help, "help", "h", false, "Show help information")
}
