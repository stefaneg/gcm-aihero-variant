package cmd

import (
	"git-clone-manager/internal/exitcodes"

	"github.com/spf13/cobra"
)

func usageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(command *cobra.Command, args []string) error {
		if err := validate(command, args); err != nil {
			return exitcodes.UsageError(err)
		}

		return nil
	}
}
