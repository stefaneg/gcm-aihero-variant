package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-clone-manager/internal/configstore"
	"git-clone-manager/internal/gitrunner"
	"git-clone-manager/internal/gitrunnertest"
)

func TestClonePassesURLToGitAsIs(t *testing.T) {
	fakeRunner := gitrunnertest.New()
	cloneRoot := t.TempDir()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rawURL := "custom+git://example.com/deep/group/repo.git"
	exitCode := executeCloneCommand(cloneRoot, fakeRunner, []string{"clone", rawURL}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	cloneCalls := fakeRunner.CloneCalls()
	if len(cloneCalls) != 1 {
		t.Fatalf("CloneCalls length = %d, want 1", len(cloneCalls))
	}

	if cloneCalls[0].URL != rawURL {
		t.Fatalf("clone URL = %q, want %q", cloneCalls[0].URL, rawURL)
	}
}

func TestCloneRejectsExistingRepositoryWithMismatchedOrigin(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	cloneRoot := t.TempDir()
	requestedURL := "https://example.com/acme/repo.git"
	conflictingOrigin := "https://example.com/other/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")
	if err := os.MkdirAll(filepath.Join(destinationPath, ".git"), 0o755); err != nil {
		t.Fatalf("create existing git repository: %v", err)
	}

	fakeRunner.SetOriginURL(conflictingOrigin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeCloneCommand(cloneRoot, fakeRunner, []string{"clone", requestedURL}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("Execute exit code = %d, want 1\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	if !strings.Contains(stderr.String(), destinationPath) || !strings.Contains(stderr.String(), conflictingOrigin) {
		t.Fatalf("stderr = %q, want destination and conflicting origin", stderr.String())
	}

	if got := fakeRunner.CloneCalls(); len(got) != 0 {
		t.Fatalf("CloneCalls = %#v, want no clone", got)
	}
}

func TestCloneRejectsExistingBrokenGitDirectoryWithActionableError(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	cloneRoot := t.TempDir()
	requestedURL := "https://example.com/acme/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")
	if err := os.MkdirAll(filepath.Join(destinationPath, ".git"), 0o755); err != nil {
		t.Fatalf("create broken git repository: %v", err)
	}

	fakeRunner.StubOriginURL(func(repoPath string) (string, error) {
		return "", errors.New("bad config")
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeCloneCommand(cloneRoot, fakeRunner, []string{"clone", requestedURL}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("Execute exit code = %d, want 1\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	for _, want := range []string{destinationPath, "destination exists but is not a git repository", "Move or remove it first"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want %q", stderr.String(), want)
		}
	}
	if strings.Contains(stderr.String(), "inspect destination origin") {
		t.Fatalf("stderr = %q, did not want low-level origin inspection error", stderr.String())
	}
	if got := fakeRunner.CloneCalls(); len(got) != 0 {
		t.Fatalf("CloneCalls = %#v, want no clone", got)
	}
}

func TestCloneAcceptsExistingRepositoryWithMatchingOrigin(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	cloneRoot := t.TempDir()
	requestedURL := "https://example.com/acme/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")
	if err := os.MkdirAll(filepath.Join(destinationPath, ".git"), 0o755); err != nil {
		t.Fatalf("create existing git repository: %v", err)
	}

	fakeRunner.SetOriginURL(requestedURL)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeCloneCommand(cloneRoot, fakeRunner, []string{"clone", requestedURL}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if stdout.String() != destinationPath+"\n" {
		t.Fatalf("stdout = %q, want destination path", stdout.String())
	}
	if got := fakeRunner.CloneCalls(); len(got) != 0 {
		t.Fatalf("CloneCalls = %#v, want no clone", got)
	}
}

func TestCloneAcceptsPreExistingEmptyDestination(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	cloneRoot := t.TempDir()
	rawURL := "https://example.com/acme/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")
	if err := os.MkdirAll(destinationPath, 0o755); err != nil {
		t.Fatalf("create empty destination: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeCloneCommand(cloneRoot, fakeRunner, []string{"clone", rawURL}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if got := stdout.String(); got != destinationPath+"\n" {
		t.Fatalf("stdout = %q, want destination path", got)
	}

	cloneCalls := fakeRunner.CloneCalls()
	if len(cloneCalls) != 1 || cloneCalls[0].DestPath != destinationPath {
		t.Fatalf("CloneCalls = %#v, want clone into existing empty destination", cloneCalls)
	}
}

func TestCloneCleansUpCreatedDestinationAfterFailure(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	cloneRoot := t.TempDir()
	rawURL := "https://example.com/acme/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")

	fakeRunner.StubClone(func(url, destPath string) error {
		if err := os.MkdirAll(filepath.Join(destPath, ".git"), 0o755); err != nil {
			t.Fatalf("simulate partial clone: %v", err)
		}
		return errors.New("clone failed")
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeCloneCommand(cloneRoot, fakeRunner, []string{"clone", rawURL}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("Execute exit code = %d, want 1\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if _, err := os.Stat(destinationPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination existence = %v, want removed", err)
	}
}

func TestCloneLeavesPreExistingEmptyDestinationAfterFailure(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	cloneRoot := t.TempDir()
	rawURL := "https://example.com/acme/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")
	if err := os.MkdirAll(destinationPath, 0o755); err != nil {
		t.Fatalf("create empty destination: %v", err)
	}

	fakeRunner.StubClone(func(url, destPath string) error {
		if err := os.WriteFile(filepath.Join(destPath, "partial"), []byte("partial\n"), 0o600); err != nil {
			t.Fatalf("simulate partial clone: %v", err)
		}
		return errors.New("clone failed")
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeCloneCommand(cloneRoot, fakeRunner, []string{"clone", rawURL}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("Execute exit code = %d, want 1\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if info, err := os.Stat(destinationPath); err != nil || !info.IsDir() {
		t.Fatalf("destination after failure = info %v err %v, want directory left in place", info, err)
	}

	if _, err := os.Stat(filepath.Join(destinationPath, "partial")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial file existence = %v, want removed from pre-existing destination", err)
	}
}

func TestMkdirAllTrackedRecordsOnlyDirectoriesItCreates(t *testing.T) {
	root := t.TempDir()
	preExisting := filepath.Join(root, "example.com")
	if err := os.Mkdir(preExisting, 0o755); err != nil {
		t.Fatalf("create pre-existing directory: %v", err)
	}

	leaf := filepath.Join(preExisting, "acme", "repo")
	created, err := mkdirAllTracked(leaf, 0o755)
	if err != nil {
		t.Fatalf("mkdirAllTracked returned error: %v", err)
	}

	want := []string{leaf, filepath.Dir(leaf)}
	if len(created) != len(want) {
		t.Fatalf("created = %#v, want %#v", created, want)
	}
	for index := range want {
		if created[index] != want[index] {
			t.Fatalf("created = %#v, want %#v", created, want)
		}
	}
}

func executeCloneCommand(cloneRoot string, runner gitrunner.Runner, args []string, stdout, stderr *bytes.Buffer) int {
	deps := DefaultDependencies()
	deps.LoadEffectiveCloneConfig = func() (configstore.EffectiveConfig, error) {
		return configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil
	}
	deps.NewGitRunner = func() gitrunner.Runner {
		return runner
	}

	return execute(args, stdout, stderr, deps)
}
