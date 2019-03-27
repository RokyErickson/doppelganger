package cmd

import (
	"github.com/spf13/cobra"
)

func Mainify(entry func(*cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(command *cobra.Command, arguments []string) {
		if err := entry(command, arguments); err != nil {
			Fatal(err)
		}
	}
}
