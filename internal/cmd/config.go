package cmd

import "github.com/spf13/cobra"

func newConfigCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "Manage gcm configuration",
	}

	command.AddCommand(newConfigSetCommand())
	command.AddCommand(newConfigShowCommand())

	return command
}

func newConfigSetCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "set",
		Short: "Set configuration values",
	}

	command.AddCommand(newConfigSetCloneRootCommand())

	return command
}

func newConfigSetCloneRootCommand() *cobra.Command {
	command := newStubCommand("clone-root <path>", "Set the clone root path")
	command.Args = usageArgs(cobra.ExactArgs(1))
	return command
}

func newConfigShowCommand() *cobra.Command {
	command := newStubCommand("show", "Show the effective configuration")
	command.Args = usageArgs(cobra.NoArgs)
	return command
}
