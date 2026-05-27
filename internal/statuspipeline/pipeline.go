package statuspipeline

import (
	"errors"
	"fmt"
	"os"

	"git-clone-manager/internal/gitrunner"
	"git-clone-manager/internal/repositorywalker"
	"git-clone-manager/internal/statuscollector"
	"git-clone-manager/internal/workerpool"
)

type Pipeline struct {
	runner gitrunner.Runner
}

func New(runner gitrunner.Runner) *Pipeline {
	return &Pipeline{runner: runner}
}

// Collect walks cloneRoot, collects status for every repository found, and
// returns all results — including those with soft errors (fetch-failed,
// no-remote). Hard errors (e.g. walker failure, unexpected git errors) are
// returned as the second value and abort collection.
func (p *Pipeline) Collect(cloneRoot string, noFetch bool) ([]statuscollector.Result, error) {
	repositories, err := repositorywalker.Walk(cloneRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("clone root %q does not exist; create it or run `gcm config set clone-root <path>` to choose an existing directory", cloneRoot)
		}
		return nil, err
	}

	collector := statuscollector.New(p.runner)
	collected := workerpool.Run(repositories, func(repositoryPath string) (statuscollector.Result, error) {
		return collector.Collect(repositoryPath, noFetch)
	})

	results := make([]statuscollector.Result, 0, len(collected))
	var batchErr error
	for _, item := range collected {
		if item.Err != nil {
			result := item.Value
			if result.RepositoryPath == "" {
				result.RepositoryPath = repositories[len(results)]
			}
			result.ErrorState = statuscollector.ErrorStateUnknown
			results = append(results, result)
			batchErr = errors.Join(batchErr, item.Err)
			continue
		}
		results = append(results, item.Value)
	}

	if batchErr != nil {
		return results, fmt.Errorf("one or more repositories could not be inspected: %w", batchErr)
	}

	return results, nil
}
