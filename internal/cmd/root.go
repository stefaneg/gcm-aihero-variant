package cmd

import (
	"git-clone-manager/internal/exitcodes"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	return newRootCommand(DefaultDependencies())
}

func newRootCommand(deps Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:           "gcm",
		Short:         "Manage cloned git repositories",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return exitcodes.UsageError(err)
	})

	root.AddCommand(newCloneCommand(deps))
	root.AddCommand(newOpenCommand(deps))
	root.AddCommand(newStatusCommand(deps))
	root.AddCommand(newConfigCommand())
	root.AddCommand(newShellInitCommand(deps))

	return root
}
