package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"git-clone-manager/internal/exitcodes"
	"git-clone-manager/internal/statuscollector"
	"git-clone-manager/internal/statusformatter"

	"github.com/spf13/cobra"
)

type statusCollector interface {
	Collect(cloneRoot string, noFetch bool) ([]statuscollector.Result, error)
}

func newStatusCommand(deps Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "status",
		Short: "Show repository status under the clone root",
		RunE: func(command *cobra.Command, args []string) error {
			effectiveConfig, err := deps.LoadEffectiveStatusConfig()
			if err != nil {
				return err
			}

			cloneRoot, err := expandHomePath(effectiveConfig.CloneRoot)
			if err != nil {
				return err
			}

			noFetch, err := command.Flags().GetBool("no-fetch")
			if err != nil {
				return err
			}

			nonDefaultOnly, err := command.Flags().GetBool("non-default")
			if err != nil {
				return err
			}

			collected, err := deps.NewStatusCollector().Collect(cloneRoot, noFetch)
			if err != nil && len(collected) == 0 {
				return err
			}
			batchErr := err

			results := make([]statuscollector.Result, 0, len(collected))
			partialFailure := false
			for _, result := range collected {
				if nonDefaultOnly && !statusResultIsNonDefault(result) {
					if result.ErrorState == statuscollector.ErrorStateFetchFailed || result.ErrorState == statuscollector.ErrorStateUnknown {
						partialFailure = true
					}
					continue
				}
				results = append(results, result)
				if result.ErrorState == statuscollector.ErrorStateFetchFailed || result.ErrorState == statuscollector.ErrorStateUnknown {
					partialFailure = true
				}
			}

			output, err := statusformatter.Format(cloneRoot, results, statusformatter.Options{
				StdoutIsTTY:          writerIsTTY(command.OutOrStdout()),
				NoColor:              os.Getenv("NO_COLOR") != "",
				NonDefaultOnly:       nonDefaultOnly,
				TotalRepositoryCount: len(collected),
			})
			if err != nil {
				return err
			}

			if _, err := fmt.Fprint(command.OutOrStdout(), output); err != nil {
				return err
			}

			if batchErr != nil {
				return batchErr
			}

			if partialFailure {
				return exitcodes.WithCode(exitcodes.General, errors.New("one or more repositories failed to fetch"))
			}

			return nil
		},
	}
	command.Args = usageArgs(cobra.NoArgs)

	command.Flags().Bool("no-fetch", false, "Use local git state without fetching remotes first")
	command.Flags().Bool("non-default", false, "Show only repositories on non-default branches")

	return command
}

func statusResultIsNonDefault(result statuscollector.Result) bool {
	if result.ErrorState == statuscollector.ErrorStateNoRemote || result.ErrorState == statuscollector.ErrorStateDefaultUnknown {
		return false
	}
	if result.DefaultBranch == "" {
		return false
	}
	return result.CurrentBranch != result.DefaultBranch
}

func writerIsTTY(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
