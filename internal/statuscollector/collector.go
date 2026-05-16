package statuscollector

import (
	"errors"

	"git-clone-manager/internal/gitrunner"
)

type ErrorState string

const (
	ErrorStateNone        ErrorState = ""
	ErrorStateFetchFailed ErrorState = "fetch-failed"
	ErrorStateNoRemote    ErrorState = "no-remote"
)

type Result struct {
	RepositoryPath string
	CurrentBranch  string
	DefaultBranch  string
	CommitsBehind  int
	DirtyCount     int
	ErrorState     ErrorState
}

type Collector struct {
	runner gitrunner.Runner
}

func New(runner gitrunner.Runner) *Collector {
	return &Collector{runner: runner}
}

func (collector *Collector) Collect(repositoryPath string, noFetch bool) (Result, error) {
	result := Result{RepositoryPath: repositoryPath}

	if !noFetch {
		if err := collector.runner.Fetch(repositoryPath); err != nil {
			switch err.(type) {
			case *gitrunner.NetworkError:
				result.ErrorState = ErrorStateFetchFailed
			case *gitrunner.NoRemoteError:
				result.ErrorState = ErrorStateNoRemote
			default:
				return Result{}, err
			}
		}
	}

	currentBranch, err := collector.runner.CurrentBranch(repositoryPath)
	if err != nil {
		return Result{}, err
	}
	result.CurrentBranch = currentBranch

	defaultBranch, err := collector.runner.DefaultBranch(repositoryPath)
	if err != nil {
		var originHeadErr *gitrunner.OriginHEADNotSetError
		var noRemoteErr *gitrunner.NoRemoteError
		switch {
		case errors.As(err, &originHeadErr):
			defaultBranch = "main"
		case errors.As(err, &noRemoteErr):
			result.ErrorState = ErrorStateNoRemote
			defaultBranch = "main"
		default:
			return Result{}, err
		}
	}
	result.DefaultBranch = defaultBranch

	if result.ErrorState == ErrorStateNoRemote {
		dirtyCount, err := collector.runner.DirtyCount(repositoryPath)
		if err != nil {
			return Result{}, err
		}
		result.DirtyCount = dirtyCount

		return result, nil
	}

	commitsBehind, err := collector.runner.CommitsBehind(repositoryPath)
	if err != nil {
		return Result{}, err
	}
	result.CommitsBehind = commitsBehind

	dirtyCount, err := collector.runner.DirtyCount(repositoryPath)
	if err != nil {
		return Result{}, err
	}
	result.DirtyCount = dirtyCount

	return result, nil
}
