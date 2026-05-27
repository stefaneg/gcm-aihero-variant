package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-clone-manager/internal/configstore"
)

func TestOpenPrintsSelectedRepositoryPath(t *testing.T) {
	cloneRoot := t.TempDir()
	firstRepository := createRepository(t, cloneRoot, "github.com", "acme", "api")
	secondRepository := createRepository(t, cloneRoot, "gitlab.com", "team", "worker")

	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	deps := stubOpenDependencies(t, configstore.EffectiveConfig{CloneRoot: cloneRoot}, func(input []string, query string, preview string) (string, int, error) {
		assertStringSlice(t, input, []string{firstRepository, secondRepository})
		if query != "" {
			t.Fatalf("query = %q, want empty", query)
		}
		if !strings.Contains(preview, "README.md") || !strings.Contains(preview, "No README found in $dir") {
			t.Fatalf("preview = %q, want README preview", preview)
		}
		return secondRepository + "\n", 0, nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute([]string{"open"}, &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if got := stdout.String(); got != secondRepository+"\n" {
		t.Fatalf("stdout = %q, want selected repository path", got)
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestOpenPassesQueryToFZF(t *testing.T) {
	cloneRoot := t.TempDir()
	selectedRepository := createRepository(t, cloneRoot, "github.com", "acme", "api")

	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	deps := stubOpenDependencies(t, configstore.EffectiveConfig{CloneRoot: cloneRoot}, func(input []string, query string, preview string) (string, int, error) {
		if query != "api" {
			t.Fatalf("query = %q, want api", query)
		}
		return selectedRepository + "\n", 0, nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute([]string{"open", "api"}, &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
}

func TestOpenAbortReturnsZeroWithEmptyStdout(t *testing.T) {
	cloneRoot := t.TempDir()
	createRepository(t, cloneRoot, "github.com", "acme", "api")

	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	deps := stubOpenDependencies(t, configstore.EffectiveConfig{CloneRoot: cloneRoot}, func(input []string, query string, preview string) (string, int, error) {
		return "", 130, nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute([]string{"open"}, &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestOpenWithMissingFZFReturnsActionableError(t *testing.T) {
	cloneRoot := t.TempDir()
	createRepository(t, cloneRoot, "github.com", "acme", "api")

	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	deps := stubOpenDependencies(t, configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil)
	deps.OpenFZFAvailable = func() bool { return false }

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute([]string{"open"}, &stdout, &stderr, deps)
	if exitCode == 0 {
		t.Fatalf("Execute exit code = %d, want nonzero", exitCode)
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "https://github.com/junegunn/fzf#installation") {
		t.Fatalf("stderr = %q, want fzf install URL", stderr.String())
	}
}

func TestOpenWithEmptyCloneRootReturnsActionableError(t *testing.T) {
	cloneRoot := t.TempDir()

	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	deps := stubOpenDependencies(t, configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute([]string{"open"}, &stdout, &stderr, deps)
	if exitCode != 1 {
		t.Fatalf("Execute exit code = %d, want 1\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), cloneRoot) || !strings.Contains(stderr.String(), "gcm clone") {
		t.Fatalf("stderr = %q, want clone root and gcm clone", stderr.String())
	}
}

func TestOpenWithMissingCloneRootReturnsDistinctActionableError(t *testing.T) {
	cloneRoot := filepath.Join(t.TempDir(), "missing")

	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	deps := stubOpenDependencies(t, configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute([]string{"open"}, &stdout, &stderr, deps)
	if exitCode != 1 {
		t.Fatalf("Execute exit code = %d, want 1\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), cloneRoot) || !strings.Contains(stderr.String(), "gcm config set clone-root") {
		t.Fatalf("stderr = %q, want clone root and config hint", stderr.String())
	}
	if strings.Contains(stderr.String(), "gcm clone") {
		t.Fatalf("stderr = %q, want distinct missing-root error", stderr.String())
	}
}

func TestOpenBareInTTYEmitsShellInitHint(t *testing.T) {
	cloneRoot := t.TempDir()
	selectedRepository := createRepository(t, cloneRoot, "github.com", "acme", "api")

	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	deps := stubOpenDependencies(t, configstore.EffectiveConfig{CloneRoot: cloneRoot}, func(input []string, query string, preview string) (string, int, error) {
		return selectedRepository + "\n", 0, nil
	})
	deps.OpenWriterIsTTY = func(writer any) bool { return true }

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute([]string{"open"}, &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), `eval "$(gcm shell-init)"`) {
		t.Fatalf("stderr = %q, want shell-init hint", stderr.String())
	}
}

func TestOpenSuppressesShellInitHintWhenStdoutIsCaptured(t *testing.T) {
	cloneRoot := t.TempDir()
	selectedRepository := createRepository(t, cloneRoot, "github.com", "acme", "api")

	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))
	deps := stubOpenDependencies(t, configstore.EffectiveConfig{CloneRoot: cloneRoot}, func(input []string, query string, preview string) (string, int, error) {
		return selectedRepository + "\n", 0, nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := execute([]string{"open"}, &stdout, &stderr, deps)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func stubOpenDependencies(t *testing.T, effectiveConfig configstore.EffectiveConfig, fzf func([]string, string, string) (string, int, error)) Dependencies {
	t.Helper()

	deps := DefaultDependencies()
	deps.LoadEffectiveOpenConfig = func() (configstore.EffectiveConfig, error) {
		return effectiveConfig, nil
	}
	if fzf != nil {
		deps.RunOpenFZF = fzf
	} else {
		deps.RunOpenFZF = func([]string, string, string) (string, int, error) {
			return "", 0, errors.New("unexpected fzf call")
		}
	}
	deps.OpenFZFAvailable = func() bool { return true }
	deps.OpenWriterIsTTY = func(writer any) bool { return false }

	return deps
}

func createRepository(t *testing.T, cloneRoot string, parts ...string) string {
	t.Helper()

	repositoryPath := filepath.Join(append([]string{cloneRoot}, parts...)...)
	if err := os.MkdirAll(filepath.Join(repositoryPath, ".git"), 0o755); err != nil {
		t.Fatalf("create repository: %v", err)
	}
	return repositoryPath
}

func assertStringSlice(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice length = %d, want %d\ngot: %#v\nwant: %#v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("slice[%d] = %q, want %q\ngot: %#v\nwant: %#v", i, got[i], want[i], got, want)
		}
	}
}
