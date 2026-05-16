package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"git-clone-manager/internal/configstore"
	"git-clone-manager/internal/repositorywalker"

	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "status",
		Short: "Show repository status under the clone root",
		RunE: func(command *cobra.Command, args []string) error {
			effectiveConfig, err := configstore.New().Effective()
			if err != nil {
				return err
			}

			cloneRoot, err := expandHomePath(effectiveConfig.CloneRoot)
			if err != nil {
				return err
			}

			if _, err := fmt.Fprintln(command.OutOrStdout(), "Repos under "+cloneRoot+":"); err != nil {
				return err
			}

			repositories, err := repositorywalker.Walk(cloneRoot)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}

			for _, repositoryPath := range repositories {
				relativePath, err := filepath.Rel(cloneRoot, repositoryPath)
				if err != nil {
					return fmt.Errorf("compute repository path relative to clone root: %w", err)
				}

				if _, err := fmt.Fprintln(command.OutOrStdout(), relativePath); err != nil {
					return err
				}
			}

			return nil
		},
	}
	command.Args = usageArgs(cobra.NoArgs)

	command.Flags().Bool("no-fetch", false, "Use local git state without fetching remotes first")
	command.Flags().Bool("non-default", false, "Show only repositories on non-default branches")

	return command
}
