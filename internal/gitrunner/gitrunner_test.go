package gitrunner_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"git-clone-manager/internal/gitrunner"
)

func TestCloneClonesRepositoryIntoDestination(t *testing.T) {
	remotePath := createBareRemote(t)
	runner := gitrunner.New()
	destinationPath := filepath.Join(t.TempDir(), "clone")

	if err := runner.Clone(remotePath, destinationPath); err != nil {
		t.Fatalf("Clone returned error: %v", err)
	}

	assertGitDirExists(t, destinationPath)
}

func TestFetchUpdatesRemoteTrackingReferences(t *testing.T) {
	remotePath := createBareRemote(t)
	clonePath := cloneRemote(t, remotePath)
	runner := gitrunner.New()

	pushCommitToRemote(t, remotePath, "second commit")

	if err := runner.Fetch(clonePath); err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	output := runGit(t, clonePath, "rev-parse", "refs/remotes/origin/main")
	if output == "" {
		t.Fatal("origin/main was not updated by fetch")
	}
}

func TestDirtyCountCountsTrackedAndUntrackedChanges(t *testing.T) {
	remotePath := createBareRemote(t)
	clonePath := cloneRemote(t, remotePath)
	runner := gitrunner.New()

	if err := os.WriteFile(filepath.Join(clonePath, "README.md"), []byte("changed\n"), 0o600); err != nil {
		t.Fatalf("write tracked change: %v", err)
	}

	if err := os.WriteFile(filepath.Join(clonePath, "notes.txt"), []byte("new file\n"), 0o600); err != nil {
		t.Fatalf("write untracked change: %v", err)
	}

	dirtyCount, err := runner.DirtyCount(clonePath)
	if err != nil {
		t.Fatalf("DirtyCount returned error: %v", err)
	}

	if dirtyCount != 2 {
		t.Fatalf("DirtyCount = %d, want %d", dirtyCount, 2)
	}
}

func TestCurrentBranchReturnsCheckedOutBranchName(t *testing.T) {
	remotePath := createBareRemote(t)
	clonePath := cloneRemote(t, remotePath)
	runner := gitrunner.New()

	runGit(t, clonePath, "checkout", "-b", "feature/status")

	currentBranch, err := runner.CurrentBranch(clonePath)
	if err != nil {
		t.Fatalf("CurrentBranch returned error: %v", err)
	}

	if currentBranch != "feature/status" {
		t.Fatalf("CurrentBranch = %q, want %q", currentBranch, "feature/status")
	}
}

func TestCommitsBehindCountsBehindRemoteDefaultBranch(t *testing.T) {
	remotePath := createBareRemote(t)
	clonePath := cloneRemote(t, remotePath)
	runner := gitrunner.New()

	pushCommitToRemote(t, remotePath, "second commit")

	if err := runner.Fetch(clonePath); err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	commitsBehind, err := runner.CommitsBehind(clonePath)
	if err != nil {
		t.Fatalf("CommitsBehind returned error: %v", err)
	}

	if commitsBehind != 1 {
		t.Fatalf("CommitsBehind = %d, want %d", commitsBehind, 1)
	}
}

func TestCommitsBehindReturnsErrorWhenOriginHEADIsUnset(t *testing.T) {
	remotePath := createBareRemote(t)
	clonePath := cloneRemote(t, remotePath)
	runner := gitrunner.New()

	pushCommitToRemote(t, remotePath, "second commit")
	if err := runner.Fetch(clonePath); err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	runGit(t, clonePath, "remote", "set-head", "origin", "--delete")

	_, err := runner.CommitsBehind(clonePath)
	if err == nil {
		t.Fatal("CommitsBehind unexpectedly succeeded")
	}

	var originHeadErr *gitrunner.OriginHEADNotSetError
	if !errors.As(err, &originHeadErr) {
		t.Fatalf("CommitsBehind error type = %T, want *gitrunner.OriginHEADNotSetError", err)
	}
}

func TestDefaultBranchReturnsOriginHEADBranchName(t *testing.T) {
	remotePath := createBareRemote(t)
	clonePath := cloneRemote(t, remotePath)
	runner := gitrunner.New()

	defaultBranch, err := runner.DefaultBranch(clonePath)
	if err != nil {
		t.Fatalf("DefaultBranch returned error: %v", err)
	}

	if defaultBranch != "main" {
		t.Fatalf("DefaultBranch = %q, want %q", defaultBranch, "main")
	}
}

func TestDefaultBranchReturnsErrorWhenOriginHEADIsUnset(t *testing.T) {
	remotePath := createBareRemote(t)
	clonePath := cloneRemote(t, remotePath)
	runner := gitrunner.New()

	runGit(t, clonePath, "remote", "set-head", "origin", "--delete")

	_, err := runner.DefaultBranch(clonePath)
	if err == nil {
		t.Fatal("DefaultBranch unexpectedly succeeded")
	}

	var originHeadErr *gitrunner.OriginHEADNotSetError
	if !errors.As(err, &originHeadErr) {
		t.Fatalf("DefaultBranch error type = %T, want *gitrunner.OriginHEADNotSetError", err)
	}
}

func TestFetchReturnsNoRemoteErrorWhenOriginIsMissing(t *testing.T) {
	repoPath := initLocalRepository(t)
	runner := gitrunner.New()

	err := runner.Fetch(repoPath)
	if err == nil {
		t.Fatal("Fetch unexpectedly succeeded")
	}

	var noRemoteErr *gitrunner.NoRemoteError
	if !errors.As(err, &noRemoteErr) {
		t.Fatalf("Fetch error type = %T, want *gitrunner.NoRemoteError", err)
	}
}

func TestCloneReturnsGitNotFoundErrorWhenGitBinaryIsMissing(t *testing.T) {
	runner := gitrunner.NewForTesting("git-does-not-exist")

	err := runner.Clone("https://example.com/repo.git", filepath.Join(t.TempDir(), "clone"))
	if err == nil {
		t.Fatal("Clone unexpectedly succeeded")
	}

	var gitNotFoundErr *gitrunner.GitNotFoundError
	if !errors.As(err, &gitNotFoundErr) {
		t.Fatalf("Clone error type = %T, want *gitrunner.GitNotFoundError", err)
	}
}

func TestFetchReturnsNetworkErrorWhenRemoteCannotBeReached(t *testing.T) {
	repoPath := initLocalRepository(t)
	runGit(t, repoPath, "remote", "add", "origin", "ssh://127.0.0.1:1/example/repo.git")
	runner := gitrunner.New()

	err := runner.Fetch(repoPath)
	if err == nil {
		t.Fatal("Fetch unexpectedly succeeded")
	}

	var networkErr *gitrunner.NetworkError
	if !errors.As(err, &networkErr) {
		t.Fatalf("Fetch error type = %T, want *gitrunner.NetworkError", err)
	}
}

func TestRepositoryOperationsReturnRepositoryNotFoundErrorForMissingPath(t *testing.T) {
	runner := gitrunner.New()

	_, err := runner.DirtyCount(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("DirtyCount unexpectedly succeeded")
	}

	var repositoryNotFoundErr *gitrunner.RepositoryNotFoundError
	if !errors.As(err, &repositoryNotFoundErr) {
		t.Fatalf("DirtyCount error type = %T, want *gitrunner.RepositoryNotFoundError", err)
	}
}

func createBareRemote(t *testing.T) string {
	t.Helper()

	remotePath := filepath.Join(t.TempDir(), "remote.git")
	runGitInDir(t, "", "init", "--bare", "--initial-branch=main", remotePath)

	worktreePath := filepath.Join(t.TempDir(), "seed")
	runGitInDir(t, "", "clone", remotePath, worktreePath)
	writeFile(t, filepath.Join(worktreePath, "README.md"), "hello\n")
	runGit(t, worktreePath, "add", "README.md")
	runGit(t, worktreePath, "commit", "-m", "initial commit")
	runGit(t, worktreePath, "push", "origin", "main")
	runGit(t, worktreePath, "remote", "set-head", "origin", "main")

	return remotePath
}

func cloneRemote(t *testing.T, remotePath string) string {
	t.Helper()

	clonePath := filepath.Join(t.TempDir(), "clone")
	runGitInDir(t, "", "clone", remotePath, clonePath)
	return clonePath
}

func pushCommitToRemote(t *testing.T, remotePath, message string) {
	t.Helper()

	worktreePath := filepath.Join(t.TempDir(), "worktree")
	runGitInDir(t, "", "clone", remotePath, worktreePath)
	writeFile(t, filepath.Join(worktreePath, "README.md"), message+"\n")
	runGit(t, worktreePath, "add", "README.md")
	runGit(t, worktreePath, "commit", "-m", message)
	runGit(t, worktreePath, "push", "origin", "main")
}

func initLocalRepository(t *testing.T) string {
	t.Helper()

	repoPath := filepath.Join(t.TempDir(), "repo")
	runGitInDir(t, "", "init", "--initial-branch=main", repoPath)
	writeFile(t, filepath.Join(repoPath, "README.md"), "hello\n")
	runGit(t, repoPath, "add", "README.md")
	runGit(t, repoPath, "commit", "-m", "initial commit")
	return repoPath
}

func assertGitDirExists(t *testing.T, repoPath string) {
	t.Helper()

	info, err := os.Stat(filepath.Join(repoPath, ".git"))
	if err != nil {
		t.Fatalf("stat .git directory: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf(".git exists but is not a directory")
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGit(t *testing.T, repoPath string, args ...string) string {
	t.Helper()
	return runGitInDir(t, repoPath, args...)
}

func runGitInDir(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Codex Test",
		"GIT_AUTHOR_EMAIL=codex@example.com",
		"GIT_COMMITTER_NAME=Codex Test",
		"GIT_COMMITTER_EMAIL=codex@example.com",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}

	return string(output)
}
