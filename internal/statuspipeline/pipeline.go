package statuspipeline

import (
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
			return nil, nil
		}
		return nil, err
	}

	collector := statuscollector.New(p.runner)
	collected := workerpool.Run(repositories, func(repositoryPath string) (statuscollector.Result, error) {
		return collector.Collect(repositoryPath, noFetch)
	})

	results := make([]statuscollector.Result, 0, len(collected))
	for _, item := range collected {
		if item.Err != nil {
			return nil, item.Err
		}
		results = append(results, item.Value)
	}

	return results, nil
}
