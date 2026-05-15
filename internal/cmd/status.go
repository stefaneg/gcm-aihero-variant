package cmd

import "github.com/spf13/cobra"

func newStatusCommand() *cobra.Command {
	command := newStubCommand("status", "Show repository status under the clone root")
	command.Args = usageArgs(cobra.NoArgs)

	command.Flags().Bool("no-fetch", false, "Use local git state without fetching remotes first")
	command.Flags().Bool("non-default", false, "Show only repositories on non-default branches")

	return command
}
