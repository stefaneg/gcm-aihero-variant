package statuscollector_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"git-clone-manager/internal/gitrunner"
	"git-clone-manager/internal/gitrunnertest"
	"git-clone-manager/internal/statuscollector"
)

func TestCollectReturnsStatusForCleanRepositoryOnDefaultBranch(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("main")
	fakeRunner.SetDefaultBranch("main")
	fakeRunner.SetCommitsBehind(0)
	fakeRunner.SetDirtyCount(0)

	collector := statuscollector.New(fakeRunner)

	result, err := collector.Collect("/repos/example", false)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if result.RepositoryPath != "/repos/example" {
		t.Fatalf("RepositoryPath = %q, want %q", result.RepositoryPath, "/repos/example")
	}

	if result.CurrentBranch != "main" {
		t.Fatalf("CurrentBranch = %q, want %q", result.CurrentBranch, "main")
	}

	if result.DefaultBranch != "main" {
		t.Fatalf("DefaultBranch = %q, want %q", result.DefaultBranch, "main")
	}

	if result.CommitsBehind != 0 {
		t.Fatalf("CommitsBehind = %d, want %d", result.CommitsBehind, 0)
	}

	if result.DirtyCount != 0 {
		t.Fatalf("DirtyCount = %d, want %d", result.DirtyCount, 0)
	}

	if result.ErrorState != statuscollector.ErrorStateNone {
		t.Fatalf("ErrorState = %q, want empty", result.ErrorState)
	}

	if got := fakeRunner.FetchCalls(); len(got) != 1 || got[0] != "/repos/example" {
		t.Fatalf("FetchCalls = %#v, want one fetch for repository", got)
	}
}

func TestCollectReturnsCheckedOutNonDefaultBranch(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("feature/status")
	fakeRunner.SetDefaultBranch("main")
	fakeRunner.SetCommitsBehind(2)
	fakeRunner.SetDirtyCount(1)

	collector := statuscollector.New(fakeRunner)

	result, err := collector.Collect("/repos/example", false)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if result.CurrentBranch != "feature/status" {
		t.Fatalf("CurrentBranch = %q, want %q", result.CurrentBranch, "feature/status")
	}

	if result.DefaultBranch != "main" {
		t.Fatalf("DefaultBranch = %q, want %q", result.DefaultBranch, "main")
	}

	if result.CommitsBehind != 2 {
		t.Fatalf("CommitsBehind = %d, want %d", result.CommitsBehind, 2)
	}

	if result.DirtyCount != 1 {
		t.Fatalf("DirtyCount = %d, want %d", result.DirtyCount, 1)
	}
}

func TestCollectReportsDefaultBranchUnknownWhenOriginHeadIsUnset(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("main")
	fakeRunner.SetDirtyCount(0)
	fakeRunner.StubDefaultBranch(func(repoPath string) (string, error) {
		return "", &gitrunner.OriginHEADNotSetError{
			RepositoryPath: repoPath,
			Err:            errors.New("origin/HEAD not set"),
		}
	})

	collector := statuscollector.New(fakeRunner)

	result, err := collector.Collect("/repos/example", false)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if result.DefaultBranch != "" {
		t.Fatalf("DefaultBranch = %q, want empty", result.DefaultBranch)
	}

	if result.ErrorState != statuscollector.ErrorStateDefaultUnknown {
		t.Fatalf("ErrorState = %q, want %q", result.ErrorState, statuscollector.ErrorStateDefaultUnknown)
	}

	if got := fakeRunner.CommitsBehindCalls(); len(got) != 0 {
		t.Fatalf("CommitsBehindCalls = %#v, want no commits-behind lookup", got)
	}
}

func TestCollectReturnsFetchFailedErrorStateWithoutFailingCollection(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("main")
	fakeRunner.SetDefaultBranch("main")
	fakeRunner.SetCommitsBehind(3)
	fakeRunner.SetDirtyCount(2)
	fakeRunner.StubFetch(func(repoPath string) error {
		return &gitrunner.NetworkError{
			Operation:      "fetch",
			RepositoryPath: repoPath,
			Err:            errors.New("host offline"),
		}
	})

	collector := statuscollector.New(fakeRunner)

	result, err := collector.Collect("/repos/example", false)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if result.ErrorState != statuscollector.ErrorStateFetchFailed {
		t.Fatalf("ErrorState = %q, want %q", result.ErrorState, statuscollector.ErrorStateFetchFailed)
	}

	if result.CurrentBranch != "main" {
		t.Fatalf("CurrentBranch = %q, want %q", result.CurrentBranch, "main")
	}

	if result.DefaultBranch != "main" {
		t.Fatalf("DefaultBranch = %q, want %q", result.DefaultBranch, "main")
	}

	if result.CommitsBehind != 3 {
		t.Fatalf("CommitsBehind = %d, want %d", result.CommitsBehind, 3)
	}

	if result.DirtyCount != 2 {
		t.Fatalf("DirtyCount = %d, want %d", result.DirtyCount, 2)
	}
}

func TestCollectClassifiesWrappedNetworkErrorAsFetchFailed(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("main")
	fakeRunner.SetDefaultBranch("main")
	fakeRunner.SetCommitsBehind(3)
	fakeRunner.SetDirtyCount(2)
	fakeRunner.StubFetch(func(repoPath string) error {
		return fmt.Errorf("fetch origin: %w", &gitrunner.NetworkError{
			Operation:      "fetch",
			RepositoryPath: repoPath,
			Err:            errors.New("host offline"),
		})
	})

	collector := statuscollector.New(fakeRunner)

	result, err := collector.Collect("/repos/example", false)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if result.ErrorState != statuscollector.ErrorStateFetchFailed {
		t.Fatalf("ErrorState = %q, want %q", result.ErrorState, statuscollector.ErrorStateFetchFailed)
	}
}

func TestCollectReturnsNoRemoteErrorStateWithoutFailingCollection(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("main")
	fakeRunner.SetDirtyCount(4)
	fakeRunner.StubFetch(func(repoPath string) error {
		return &gitrunner.NoRemoteError{
			RepositoryPath: repoPath,
			Err:            errors.New("origin missing"),
		}
	})
	fakeRunner.StubDefaultBranch(func(repoPath string) (string, error) {
		return "", &gitrunner.NoRemoteError{
			RepositoryPath: repoPath,
			Err:            errors.New("origin missing"),
		}
	})

	collector := statuscollector.New(fakeRunner)

	result, err := collector.Collect("/repos/example", false)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if result.ErrorState != statuscollector.ErrorStateNoRemote {
		t.Fatalf("ErrorState = %q, want %q", result.ErrorState, statuscollector.ErrorStateNoRemote)
	}

	if result.CurrentBranch != "main" {
		t.Fatalf("CurrentBranch = %q, want %q", result.CurrentBranch, "main")
	}

	if result.DefaultBranch != "" {
		t.Fatalf("DefaultBranch = %q, want empty", result.DefaultBranch)
	}

	if result.CommitsBehind != 0 {
		t.Fatalf("CommitsBehind = %d, want %d", result.CommitsBehind, 0)
	}

	if result.DirtyCount != 4 {
		t.Fatalf("DirtyCount = %d, want %d", result.DirtyCount, 4)
	}

	if got := fakeRunner.CommitsBehindCalls(); len(got) != 0 {
		t.Fatalf("CommitsBehindCalls = %#v, want no commits-behind lookup", got)
	}
}

func TestCollectSkipsFetchWhenNoFetchIsEnabled(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("main")
	fakeRunner.SetDefaultBranch("main")
	fakeRunner.SetCommitsBehind(5)
	fakeRunner.SetDirtyCount(0)

	collector := statuscollector.New(fakeRunner)

	result, err := collector.Collect("/repos/example", true)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	if result.CommitsBehind != 5 {
		t.Fatalf("CommitsBehind = %d, want %d", result.CommitsBehind, 5)
	}

	if got := fakeRunner.FetchCalls(); len(got) != 0 {
		t.Fatalf("FetchCalls = %#v, want no fetch calls", got)
	}
}

func TestCollectSupportsConcurrentUseWithSharedRunner(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("main")
	fakeRunner.SetDefaultBranch("main")
	fakeRunner.SetCommitsBehind(1)
	fakeRunner.SetDirtyCount(0)

	collector := statuscollector.New(fakeRunner)

	var waitGroup sync.WaitGroup
	errs := make(chan error, 20)
	for index := range 20 {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()

			result, err := collector.Collect(fmt.Sprintf("/repos/example-%d", index), false)
			if err != nil {
				errs <- err
				return
			}

			if result.CurrentBranch != "main" {
				errs <- fmt.Errorf("CurrentBranch = %q, want %q", result.CurrentBranch, "main")
			}
		}(index)
	}

	waitGroup.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("Collect returned error: %v", err)
	}

	if got := len(fakeRunner.FetchCalls()); got != 20 {
		t.Fatalf("FetchCalls length = %d, want %d", got, 20)
	}
}
