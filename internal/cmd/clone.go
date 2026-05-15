package cmd

import "github.com/spf13/cobra"

func newCloneCommand() *cobra.Command {
	command := newStubCommand("clone <url>", "Clone a repository into its derived path")
	command.Args = usageArgs(cobra.ExactArgs(1))
	return command
}
