package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RokyErickson/doppelganger/cmd"
	"github.com/RokyErickson/doppelganger/pkg/agent"
	"github.com/RokyErickson/doppelganger/pkg/doppelganger"
)

func legalMain(command *cobra.Command, arguments []string) error {
	fmt.Println(doppelganger.LegalNotice)

	return nil
}

var legalCommand = &cobra.Command{
	Use:   agent.ModeLegal,
	Short: "Show legal information",
	Run:   cmd.Mainify(legalMain),
}

var legalConfiguration struct {
	help bool
}

func init() {
	flags := legalCommand.Flags()
	flags.BoolVarP(&legalConfiguration.help, "help", "h", false, "Show help information")
}
