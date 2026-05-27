package statuspipeline_test

import (
	"errors"
	"git-clone-manager/internal/gitrunner/fake"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-clone-manager/internal/statuscollector"
	"git-clone-manager/internal/statuspipeline"
)

func TestCollectReturnsPartialResultsWhenOneRepositoryHasHardError(t *testing.T) {
	cloneRoot := t.TempDir()
	goodRepo := filepath.Join(cloneRoot, "github.com", "acme", "good")
	badRepo := filepath.Join(cloneRoot, "github.com", "acme", "bad")
	mkdirGitDir(t, goodRepo)
	mkdirGitDir(t, badRepo)

	runner := fake.New()
	runner.SetCurrentBranch("main")
	runner.SetDefaultBranch("main")
	runner.StubCurrentBranch(func(repoPath string) (string, error) {
		if repoPath == badRepo {
			return "", errors.New("corrupt git metadata")
		}
		return "main", nil
	})

	results, err := statuspipeline.New(runner).Collect(cloneRoot, false)
	if err == nil {
		t.Fatal("Collect error = nil, want batch error")
	}

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2: %#v", len(results), results)
	}

	var sawGood bool
	var sawBad bool
	for _, result := range results {
		switch result.RepositoryPath {
		case goodRepo:
			sawGood = true
			if result.ErrorState != statuscollector.ErrorStateNone {
				t.Fatalf("good repo ErrorState = %q, want empty", result.ErrorState)
			}
		case badRepo:
			sawBad = true
			if result.ErrorState != statuscollector.ErrorStateUnknown {
				t.Fatalf("bad repo ErrorState = %q, want %q", result.ErrorState, statuscollector.ErrorStateUnknown)
			}
		}
	}
	if !sawGood || !sawBad {
		t.Fatalf("results = %#v, want good and bad repos", results)
	}
}

func TestCollectReturnsActionableErrorWhenCloneRootIsMissing(t *testing.T) {
	cloneRoot := filepath.Join(t.TempDir(), "missing")

	results, err := statuspipeline.New(fake.New()).Collect(cloneRoot, false)
	if err == nil {
		t.Fatal("Collect error = nil, want missing clone root error")
	}
	if results != nil {
		t.Fatalf("results = %#v, want nil", results)
	}

	message := err.Error()
	for _, want := range []string{cloneRoot, "gcm config set clone-root"} {
		if !strings.Contains(message, want) {
			t.Fatalf("error = %q, want %q", message, want)
		}
	}
}

func TestCollectReturnsEmptyResultsForExistingEmptyCloneRoot(t *testing.T) {
	cloneRoot := t.TempDir()

	results, err := statuspipeline.New(fake.New()).Collect(cloneRoot, false)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}

func mkdirGitDir(t *testing.T, repoPath string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0o755); err != nil {
		t.Fatalf("create git dir: %v", err)
	}
}
