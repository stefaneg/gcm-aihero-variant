package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStubCommand(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(command *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(command.OutOrStdout(), "not yet implemented")
			return err
		},
	}
}
