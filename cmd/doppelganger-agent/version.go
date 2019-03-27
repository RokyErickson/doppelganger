package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/agent"
	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
)

func versionMain(command *cobra.Command, arguments []string) error {

	fmt.Println(doppelganger.Version)

	return nil
}

var versionCommand = &cobra.Command{
	Use:   agent.ModeVersion,
	Short: "Show version information",
	Run:   cmd.Mainify(versionMain),
}

var versionConfiguration struct {
	help bool
}

func init() {

	flags := versionCommand.Flags()

	flags.BoolVarP(&versionConfiguration.help, "help", "h", false, "Show help information")
}
