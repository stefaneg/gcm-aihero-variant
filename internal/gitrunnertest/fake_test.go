package gitrunnertest_test

import (
	"errors"
	"sync"
	"testing"

	"git-clone-manager/internal/gitrunnertest"
)

func TestFakeReturnsStubbedValuesAndRecordsCalls(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetCurrentBranch("main")
	fakeRunner.SetDirtyCount(3)
	fakeRunner.SetCommitsBehind(2)
	fakeRunner.SetDefaultBranch("main")

	cloneErr := errors.New("clone failed")
	fetchErr := errors.New("fetch failed")
	currentBranchErr := errors.New("current branch failed")
	fakeRunner.StubClone(func(url, destPath string) error {
		return cloneErr
	})
	fakeRunner.StubFetch(func(repoPath string) error {
		return fetchErr
	})
	fakeRunner.StubCurrentBranch(func(repoPath string) (string, error) {
		return "", currentBranchErr
	})

	if err := fakeRunner.Clone("https://example.com/repo.git", "/tmp/repo"); !errors.Is(err, cloneErr) {
		t.Fatalf("Clone error = %v, want %v", err, cloneErr)
	}

	if err := fakeRunner.Fetch("/tmp/repo"); !errors.Is(err, fetchErr) {
		t.Fatalf("Fetch error = %v, want %v", err, fetchErr)
	}

	if _, err := fakeRunner.CurrentBranch("/tmp/repo"); !errors.Is(err, currentBranchErr) {
		t.Fatalf("CurrentBranch error = %v, want %v", err, currentBranchErr)
	}

	dirtyCount, err := fakeRunner.DirtyCount("/tmp/repo")
	if err != nil {
		t.Fatalf("DirtyCount returned error: %v", err)
	}

	if dirtyCount != 3 {
		t.Fatalf("DirtyCount = %d, want %d", dirtyCount, 3)
	}

	commitsBehind, err := fakeRunner.CommitsBehind("/tmp/repo")
	if err != nil {
		t.Fatalf("CommitsBehind returned error: %v", err)
	}

	if commitsBehind != 2 {
		t.Fatalf("CommitsBehind = %d, want %d", commitsBehind, 2)
	}

	defaultBranch, err := fakeRunner.DefaultBranch("/tmp/repo")
	if err != nil {
		t.Fatalf("DefaultBranch returned error: %v", err)
	}

	if defaultBranch != "main" {
		t.Fatalf("DefaultBranch = %q, want %q", defaultBranch, "main")
	}

	if got := fakeRunner.CloneCalls(); len(got) != 1 || got[0].URL != "https://example.com/repo.git" || got[0].DestPath != "/tmp/repo" {
		t.Fatalf("CloneCalls = %#v, want one recorded clone call", got)
	}

	if got := fakeRunner.FetchCalls(); len(got) != 1 || got[0] != "/tmp/repo" {
		t.Fatalf("FetchCalls = %#v, want one recorded fetch call", got)
	}

	if got := fakeRunner.CurrentBranchCalls(); len(got) != 1 || got[0] != "/tmp/repo" {
		t.Fatalf("CurrentBranchCalls = %#v, want one recorded current-branch call", got)
	}

	if got := fakeRunner.DirtyCountCalls(); len(got) != 1 || got[0] != "/tmp/repo" {
		t.Fatalf("DirtyCountCalls = %#v, want one recorded dirty-count call", got)
	}

	if got := fakeRunner.CommitsBehindCalls(); len(got) != 1 || got[0] != "/tmp/repo" {
		t.Fatalf("CommitsBehindCalls = %#v, want one recorded commits-behind call", got)
	}

	if got := fakeRunner.DefaultBranchCalls(); len(got) != 1 || got[0] != "/tmp/repo" {
		t.Fatalf("DefaultBranchCalls = %#v, want one recorded default-branch call", got)
	}
}

func TestFakeSupportsConcurrentUse(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	fakeRunner.SetDirtyCount(1)

	var waitGroup sync.WaitGroup
	for range 20 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			if _, err := fakeRunner.DirtyCount("/tmp/repo"); err != nil {
				t.Errorf("DirtyCount returned error: %v", err)
			}
		}()
	}

	waitGroup.Wait()

	if got := len(fakeRunner.DirtyCountCalls()); got != 20 {
		t.Fatalf("DirtyCountCalls length = %d, want %d", got, 20)
	}
}
