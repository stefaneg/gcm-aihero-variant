package gitrunner

import (
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

type gitCall struct {
	gitBinary string
	repoPath  string
	args      []string
}

func newFakeRunner(fn func(repoPath string, args ...string) (string, error)) (*runner, *[]gitCall) {
	var calls []gitCall
	return &runner{
		gitBinary: "git-test",
		runCommand: func(gitBinary, repoPath string, args ...string) (string, error) {
			calls = append(calls, gitCall{
				gitBinary: gitBinary,
				repoPath:  repoPath,
				args:      append([]string(nil), args...),
			})
			return fn(repoPath, args...)
		},
	}, &calls
}

func TestCloneRunsGitCloneIntoDestination(t *testing.T) {
	runner, calls := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return "", nil
	})

	if err := runner.Clone("https://example.com/repo.git", "/repos/example"); err != nil {
		t.Fatalf("Clone returned error: %v", err)
	}

	assertGitCall(t, *calls, gitCall{
		gitBinary: "git-test",
		args:      []string{"clone", "https://example.com/repo.git", "/repos/example"},
	})
}

func TestFetchRunsGitFetchOriginInRepository(t *testing.T) {
	runner, calls := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return "", nil
	})

	if err := runner.Fetch("/repos/example"); err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	assertGitCall(t, *calls, gitCall{
		gitBinary: "git-test",
		repoPath:  "/repos/example",
		args:      []string{"fetch", "origin"},
	})
}

func TestDirtyCountCountsPorcelainRows(t *testing.T) {
	runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return " M README.md\n?? notes.txt\n\n", nil
	})

	dirtyCount, err := runner.DirtyCount("/repos/example")
	if err != nil {
		t.Fatalf("DirtyCount returned error: %v", err)
	}

	if dirtyCount != 2 {
		t.Fatalf("DirtyCount = %d, want %d", dirtyCount, 2)
	}
}

func TestCurrentBranchReturnsCheckedOutBranchName(t *testing.T) {
	runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return "feature/status\n", nil
	})

	currentBranch, err := runner.CurrentBranch("/repos/example")
	if err != nil {
		t.Fatalf("CurrentBranch returned error: %v", err)
	}

	if currentBranch != "feature/status" {
		t.Fatalf("CurrentBranch = %q, want %q", currentBranch, "feature/status")
	}
}

func TestCommitsBehindCountsBehindRemoteDefaultBranch(t *testing.T) {
	runner, calls := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		switch args[0] {
		case "symbolic-ref":
			return "origin/main\n", nil
		case "rev-list":
			return "3\n", nil
		default:
			t.Fatalf("unexpected git args: %v", args)
			return "", nil
		}
	})

	commitsBehind, err := runner.CommitsBehind("/repos/example")
	if err != nil {
		t.Fatalf("CommitsBehind returned error: %v", err)
	}

	if commitsBehind != 3 {
		t.Fatalf("CommitsBehind = %d, want %d", commitsBehind, 3)
	}
	if got := (*calls)[1].args; !reflect.DeepEqual(got, []string{"rev-list", "--count", "refs/heads/main..refs/remotes/origin/main"}) {
		t.Fatalf("rev-list args = %#v, want default branch comparison", got)
	}
}

func TestCommitsBehindReturnsErrorWhenOriginHEADIsUnset(t *testing.T) {
	runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return "", &commandError{
			stderr: "fatal: ref refs/remotes/origin/HEAD is not a symbolic ref",
			err:    &exec.ExitError{},
		}
	})

	_, err := runner.CommitsBehind("/repos/example")
	if err == nil {
		t.Fatal("CommitsBehind unexpectedly succeeded")
	}

	var originHeadErr *OriginHEADNotSetError
	if !errors.As(err, &originHeadErr) {
		t.Fatalf("CommitsBehind error type = %T, want *OriginHEADNotSetError", err)
	}
}

func TestDefaultBranchReturnsOriginHEADBranchName(t *testing.T) {
	runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return "origin/main\n", nil
	})

	defaultBranch, err := runner.DefaultBranch("/repos/example")
	if err != nil {
		t.Fatalf("DefaultBranch returned error: %v", err)
	}

	if defaultBranch != "main" {
		t.Fatalf("DefaultBranch = %q, want %q", defaultBranch, "main")
	}
}

func TestDefaultBranchRejectsRefOutsideOrigin(t *testing.T) {
	runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return "upstream/main\n", nil
	})

	_, err := runner.DefaultBranch("/repos/example")
	if err == nil {
		t.Fatal("DefaultBranch unexpectedly succeeded")
	}

	if !strings.Contains(err.Error(), "does not point at origin/<branch>") {
		t.Fatalf("DefaultBranch error = %v, want invalid origin HEAD message", err)
	}
}

func TestOriginURLReturnsConfiguredOrigin(t *testing.T) {
	runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return "https://example.com/acme/repo.git\n", nil
	})

	originURL, err := runner.OriginURL("/repos/example")
	if err != nil {
		t.Fatalf("OriginURL returned error: %v", err)
	}

	if originURL != "https://example.com/acme/repo.git" {
		t.Fatalf("OriginURL = %q, want configured URL", originURL)
	}
}

func TestClassifiesGitErrors(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		stderr     string
		assertType func(error) bool
	}{
		{
			name:      "repository missing",
			operation: "status",
			stderr:    "fatal: cannot change to '/missing': No such file or directory",
			assertType: func(err error) bool {
				var target *RepositoryNotFoundError
				return errors.As(err, &target)
			},
		},
		{
			name:      "origin head not set",
			operation: "symbolic-ref",
			stderr:    "fatal: ref refs/remotes/origin/HEAD is not a symbolic ref",
			assertType: func(err error) bool {
				var target *OriginHEADNotSetError
				return errors.As(err, &target)
			},
		},
		{
			name:      "no remote",
			operation: "fetch",
			stderr:    "fatal: no remote repository specified",
			assertType: func(err error) bool {
				var target *NoRemoteError
				return errors.As(err, &target)
			},
		},
		{
			name:      "network",
			operation: "fetch",
			stderr:    "fatal: could not read from remote repository",
			assertType: func(err error) bool {
				var target *NetworkError
				return errors.As(err, &target)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
				return "", &commandError{stderr: test.stderr, err: &exec.ExitError{}}
			})

			err := runner.classifyError(test.operation, "/repos/example", &commandError{stderr: test.stderr, err: &exec.ExitError{}})
			if err == nil {
				t.Fatal("classifyError unexpectedly returned nil")
			}

			if !test.assertType(err) {
				t.Fatalf("error type = %T, error = %v", err, err)
			}
		})
	}
}

func TestCloneReturnsGitNotFoundErrorWhenGitBinaryIsMissing(t *testing.T) {
	runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		return "", exec.ErrNotFound
	})

	err := runner.Clone("https://example.com/repo.git", "/repos/example")
	if err == nil {
		t.Fatal("Clone unexpectedly succeeded")
	}

	var gitNotFoundErr *GitNotFoundError
	if !errors.As(err, &gitNotFoundErr) {
		t.Fatalf("Clone error type = %T, want *GitNotFoundError", err)
	}
}

func TestCommitsBehindReturnsParseErrorForInvalidCount(t *testing.T) {
	runner, _ := newFakeRunner(func(repoPath string, args ...string) (string, error) {
		switch args[0] {
		case "symbolic-ref":
			return "origin/main\n", nil
		case "rev-list":
			return "not-a-number\n", nil
		default:
			t.Fatalf("unexpected git args: %v", args)
			return "", nil
		}
	})

	_, err := runner.CommitsBehind("/repos/example")
	if err == nil {
		t.Fatal("CommitsBehind unexpectedly succeeded")
	}

	if !strings.Contains(err.Error(), "parse commits behind count") {
		t.Fatalf("CommitsBehind error = %v, want parse error", err)
	}
}

func assertGitCall(t *testing.T, calls []gitCall, want gitCall) {
	t.Helper()

	if len(calls) != 1 {
		t.Fatalf("git calls = %#v, want one call", calls)
	}
	if calls[0].gitBinary != want.gitBinary || calls[0].repoPath != want.repoPath || !reflect.DeepEqual(calls[0].args, want.args) {
		t.Fatalf("git call = %#v, want %#v", calls[0], want)
	}
}
