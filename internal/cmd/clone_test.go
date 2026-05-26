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

	originalLoadConfig := loadEffectiveCloneConfig
	originalNewGitRunner := newGitRunner
	t.Cleanup(func() {
		loadEffectiveCloneConfig = originalLoadConfig
		newGitRunner = originalNewGitRunner
	})

	loadEffectiveCloneConfig = func() (configstore.EffectiveConfig, error) {
		return configstore.EffectiveConfig{CloneRoot: t.TempDir()}, nil
	}
	newGitRunner = func() gitrunner.Runner {
		return fakeRunner
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rawURL := "custom+git://example.com/deep/group/repo.git"
	exitCode := Execute([]string{"clone", rawURL}, &stdout, &stderr)
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

	originalLoadConfig := loadEffectiveCloneConfig
	originalNewGitRunner := newGitRunner
	t.Cleanup(func() {
		loadEffectiveCloneConfig = originalLoadConfig
		newGitRunner = originalNewGitRunner
	})

	cloneRoot := t.TempDir()
	requestedURL := "https://example.com/acme/repo.git"
	conflictingOrigin := "https://example.com/other/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")
	if err := os.MkdirAll(filepath.Join(destinationPath, ".git"), 0o755); err != nil {
		t.Fatalf("create existing git repository: %v", err)
	}

	loadEffectiveCloneConfig = func() (configstore.EffectiveConfig, error) {
		return configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil
	}
	fakeRunner.SetOriginURL(conflictingOrigin)
	newGitRunner = func() gitrunner.Runner {
		return fakeRunner
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"clone", requestedURL}, &stdout, &stderr)
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

func TestCloneAcceptsPreExistingEmptyDestination(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	originalLoadConfig := loadEffectiveCloneConfig
	originalNewGitRunner := newGitRunner
	t.Cleanup(func() {
		loadEffectiveCloneConfig = originalLoadConfig
		newGitRunner = originalNewGitRunner
	})

	cloneRoot := t.TempDir()
	rawURL := "https://example.com/acme/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")
	if err := os.MkdirAll(destinationPath, 0o755); err != nil {
		t.Fatalf("create empty destination: %v", err)
	}

	loadEffectiveCloneConfig = func() (configstore.EffectiveConfig, error) {
		return configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil
	}
	newGitRunner = func() gitrunner.Runner {
		return fakeRunner
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"clone", rawURL}, &stdout, &stderr)
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

	originalLoadConfig := loadEffectiveCloneConfig
	originalNewGitRunner := newGitRunner
	t.Cleanup(func() {
		loadEffectiveCloneConfig = originalLoadConfig
		newGitRunner = originalNewGitRunner
	})

	cloneRoot := t.TempDir()
	rawURL := "https://example.com/acme/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")

	loadEffectiveCloneConfig = func() (configstore.EffectiveConfig, error) {
		return configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil
	}
	fakeRunner.StubClone(func(url, destPath string) error {
		if err := os.MkdirAll(filepath.Join(destPath, ".git"), 0o755); err != nil {
			t.Fatalf("simulate partial clone: %v", err)
		}
		return errors.New("clone failed")
	})
	newGitRunner = func() gitrunner.Runner {
		return fakeRunner
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"clone", rawURL}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("Execute exit code = %d, want 1\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if _, err := os.Stat(destinationPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination existence = %v, want removed", err)
	}
}

func TestCloneLeavesPreExistingEmptyDestinationAfterFailure(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	originalLoadConfig := loadEffectiveCloneConfig
	originalNewGitRunner := newGitRunner
	t.Cleanup(func() {
		loadEffectiveCloneConfig = originalLoadConfig
		newGitRunner = originalNewGitRunner
	})

	cloneRoot := t.TempDir()
	rawURL := "https://example.com/acme/repo.git"
	destinationPath := filepath.Join(cloneRoot, "example.com", "acme", "repo")
	if err := os.MkdirAll(destinationPath, 0o755); err != nil {
		t.Fatalf("create empty destination: %v", err)
	}

	loadEffectiveCloneConfig = func() (configstore.EffectiveConfig, error) {
		return configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil
	}
	fakeRunner.StubClone(func(url, destPath string) error {
		if err := os.WriteFile(filepath.Join(destPath, "partial"), []byte("partial\n"), 0o600); err != nil {
			t.Fatalf("simulate partial clone: %v", err)
		}
		return errors.New("clone failed")
	})
	newGitRunner = func() gitrunner.Runner {
		return fakeRunner
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"clone", rawURL}, &stdout, &stderr)
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
