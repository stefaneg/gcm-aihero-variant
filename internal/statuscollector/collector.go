package statuscollector

import (
	"context"
	"errors"

	"git-clone-manager/internal/gitrunner"
)

type ErrorState string

const (
	ErrorStateNone           ErrorState = ""
	ErrorStateFetchFailed    ErrorState = "fetch-failed"
	ErrorStateNoRemote       ErrorState = "no-remote"
	ErrorStateDefaultUnknown ErrorState = "default-unknown"
	ErrorStateUnknown        ErrorState = "error"
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
	ctx, cancel := context.WithTimeout(context.Background(), gitrunner.DefaultCommandTimeout)
	defer cancel()

	if !noFetch {
		if err := collector.runner.Fetch(ctx, repositoryPath); err != nil {
			var networkErr *gitrunner.NetworkError
			var noRemoteErr *gitrunner.NoRemoteError
			var timeoutErr *gitrunner.TimeoutError
			switch {
			case errors.As(err, &networkErr):
				result.ErrorState = ErrorStateFetchFailed
			case errors.As(err, &timeoutErr):
				result.ErrorState = ErrorStateFetchFailed
			case errors.As(err, &noRemoteErr):
				result.ErrorState = ErrorStateNoRemote
			default:
				return Result{}, err
			}
		}
	}

	currentBranch, err := collector.runner.CurrentBranch(ctx, repositoryPath)
	if err != nil {
		return Result{}, err
	}
	result.CurrentBranch = currentBranch

	defaultBranch, err := collector.runner.DefaultBranch(ctx, repositoryPath)
	if err != nil {
		var originHeadErr *gitrunner.OriginHEADNotSetError
		var noRemoteErr *gitrunner.NoRemoteError
		switch {
		case errors.As(err, &originHeadErr):
			if result.ErrorState != ErrorStateNoRemote {
				result.ErrorState = ErrorStateDefaultUnknown
			}
		case errors.As(err, &noRemoteErr):
			result.ErrorState = ErrorStateNoRemote
		default:
			return Result{}, err
		}
	}
	result.DefaultBranch = defaultBranch

	if result.ErrorState == ErrorStateNoRemote || result.ErrorState == ErrorStateDefaultUnknown {
		dirtyCount, err := collector.runner.DirtyCount(ctx, repositoryPath)
		if err != nil {
			return Result{}, err
		}
		result.DirtyCount = dirtyCount

		return result, nil
	}

	commitsBehind, err := collector.runner.CommitsBehind(ctx, repositoryPath)
	if err != nil {
		return Result{}, err
	}
	result.CommitsBehind = commitsBehind

	dirtyCount, err := collector.runner.DirtyCount(ctx, repositoryPath)
	if err != nil {
		return Result{}, err
	}
	result.DirtyCount = dirtyCount

	return result, nil
}
