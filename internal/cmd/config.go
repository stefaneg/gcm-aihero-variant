package cmd

import (
	"fmt"

	"git-clone-manager/internal/configstore"

	"github.com/spf13/cobra"
)

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
	command.AddCommand(newConfigSetProjectOpenerCommand())

	return command
}

func newConfigSetCloneRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "clone-root <path>",
		Short: "Set the clone root path",
		RunE: func(command *cobra.Command, args []string) error {
			configPath, err := configstore.New().SetCloneRoot(args[0])
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(command.OutOrStdout(), "Config saved to "+configPath)
			return err
		},
	}
	command.Args = usageArgs(cobra.ExactArgs(1))
	return command
}

func newConfigSetProjectOpenerCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "project-opener <command>",
		Short: "Set the project opener command",
		RunE: func(command *cobra.Command, args []string) error {
			configPath, err := configstore.New().SetProjectOpener(args[0])
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(command.OutOrStdout(), "Config saved to "+configPath); err != nil {
				return err
			}

			_, err = fmt.Fprintln(command.ErrOrStderr(), `Project opener changes take effect after re-running: eval "$(gcm shell-init)"`)
			return err
		},
	}
	command.Args = usageArgs(cobra.ExactArgs(1))
	return command
}

func newConfigShowCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "show",
		Short: "Show the effective configuration",
		RunE: func(command *cobra.Command, args []string) error {
			effectiveConfig, err := configstore.New().Effective()
			if err != nil {
				return err
			}

			line := "clone_root: " + effectiveConfig.CloneRoot
			if effectiveConfig.CloneRootIsDefault {
				line += "  # default"
			}

			if _, err := fmt.Fprintln(command.OutOrStdout(), line); err != nil {
				return err
			}

			projectOpener := effectiveConfig.ProjectOpener
			if projectOpener == "" {
				projectOpener = `""`
			}
			line = "project_opener: " + projectOpener
			if effectiveConfig.ProjectOpenerIsDefault {
				line += "  # default"
			}

			_, err = fmt.Fprintln(command.OutOrStdout(), line)
			return err
		},
	}
	command.Args = usageArgs(cobra.NoArgs)
	return command
}
