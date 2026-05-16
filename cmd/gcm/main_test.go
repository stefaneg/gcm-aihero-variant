package main_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git-clone-manager/internal/derivedpath"
	"git-clone-manager/internal/exitcodes"
	"git-clone-manager/internal/repourl"
)

func TestHelpListsTopLevelCommands(t *testing.T) {
	binary := buildGCM(t)

	cmd := exec.Command(binary, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm --help failed: %v\n%s", err, output)
	}

	help := string(output)
	for _, want := range []string{"clone", "status", "config"} {
		if !strings.Contains(help, want) {
			t.Fatalf("gcm --help missing %q in output:\n%s", want, help)
		}
	}
}

func TestCommandHelpRendersUsage(t *testing.T) {
	binary := buildGCM(t)

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "clone help",
			args: []string{"clone", "--help"},
			want: []string{"Usage:", "clone <url>"},
		},
		{
			name: "config help",
			args: []string{"config", "--help"},
			want: []string{"Usage:", "set", "show"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := exec.Command(binary, test.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s failed: %v\n%s", strings.Join(test.args, " "), err, output)
			}

			help := string(output)
			for _, want := range test.want {
				if !strings.Contains(help, want) {
					t.Fatalf("%s missing %q in output:\n%s", strings.Join(test.args, " "), want, help)
				}
			}
		})
	}
}

func TestStatusHelpDocumentsFlags(t *testing.T) {
	binary := buildGCM(t)

	cmd := exec.Command(binary, "status", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm status --help failed: %v\n%s", err, output)
	}

	help := string(output)
	for _, want := range []string{"--no-fetch", "--non-default"} {
		if !strings.Contains(help, want) {
			t.Fatalf("gcm status --help missing %q in output:\n%s", want, help)
		}
	}
}

func TestConfigSetHelpDocumentsCloneRootSubcommand(t *testing.T) {
	binary := buildGCM(t)

	cmd := exec.Command(binary, "config", "set", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm config set --help failed: %v\n%s", err, output)
	}

	help := string(output)
	if !strings.Contains(help, "clone-root") {
		t.Fatalf("gcm config set --help missing %q in output:\n%s", "clone-root", help)
	}
}

func TestConfigShowHelpRendersUsage(t *testing.T) {
	binary := buildGCM(t)

	cmd := exec.Command(binary, "config", "show", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm config show --help failed: %v\n%s", err, output)
	}

	help := string(output)
	for _, want := range []string{"Usage:", "show"} {
		if !strings.Contains(help, want) {
			t.Fatalf("gcm config show --help missing %q in output:\n%s", want, help)
		}
	}
}

func TestCloneClonesRepositoryIntoDerivedPath(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")
	remotePath := createBareRemote(t)
	remoteURL := "file://localhost" + remotePath

	parts, err := repourl.Parse(remoteURL)
	if err != nil {
		t.Fatalf("parse remote URL for expected path: %v", err)
	}

	wantPath := derivedpath.Derive(cloneRoot, parts.Hostname, parts.PathPrefix, parts.RepositoryName)

	cmd := exec.Command(binary, "clone", remoteURL)
	cmd.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm clone failed: %v\n%s", err, output)
	}

	cloningLine := "Cloning to " + wantPath + "..."
	if !strings.Contains(string(output), cloningLine) {
		t.Fatalf("gcm clone output = %q, want it to mention %q", output, cloningLine)
	}

	if strings.Index(string(output), cloningLine) > strings.Index(string(output), "Done.") {
		t.Fatalf("gcm clone output = %q, want derived path announcement before completion", output)
	}

	assertGitDirExists(t, wantPath)
}

func TestCloneWarnsAndCreatesMissingCloneRoot(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "missing", "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")
	remoteURL := "file://localhost" + createBareRemote(t)

	cmd := exec.Command(binary, "clone", remoteURL)
	cmd.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm clone failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "Clone root "+cloneRoot+" does not exist - creating it") {
		t.Fatalf("gcm clone output = %q, want clone-root creation warning", output)
	}

	info, err := os.Stat(cloneRoot)
	if err != nil {
		t.Fatalf("stat clone root: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf("clone root exists but is not a directory")
	}
}

func TestCloneReportsAlreadyClonedDestination(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")
	remoteURL := "file://localhost" + createBareRemote(t)

	firstClone := exec.Command(binary, "clone", remoteURL)
	firstClone.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	if output, err := firstClone.CombinedOutput(); err != nil {
		t.Fatalf("initial gcm clone failed: %v\n%s", err, output)
	}

	parts, err := repourl.Parse(remoteURL)
	if err != nil {
		t.Fatalf("parse remote URL for expected path: %v", err)
	}

	wantPath := derivedpath.Derive(cloneRoot, parts.Hostname, parts.PathPrefix, parts.RepositoryName)

	secondClone := exec.Command(binary, "clone", remoteURL)
	secondClone.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := secondClone.CombinedOutput()
	if err != nil {
		t.Fatalf("second gcm clone failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), "Already cloned at "+wantPath) {
		t.Fatalf("second gcm clone output = %q, want already-cloned message for %q", output, wantPath)
	}
}

func TestCloneReturnsActionableErrorWhenDestinationIsNotAGitRepository(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")
	remoteURL := "file://localhost" + createBareRemote(t)

	parts, err := repourl.Parse(remoteURL)
	if err != nil {
		t.Fatalf("parse remote URL for expected path: %v", err)
	}

	destinationPath := derivedpath.Derive(cloneRoot, parts.Hostname, parts.PathPrefix, parts.RepositoryName)
	if err := os.MkdirAll(destinationPath, 0o755); err != nil {
		t.Fatalf("create destination directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(destinationPath, "README.md"), []byte("not a repo\n"), 0o600); err != nil {
		t.Fatalf("write destination file: %v", err)
	}

	cmd := exec.Command(binary, "clone", remoteURL)
	cmd.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("gcm clone unexpectedly succeeded:\n%s", output)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exit error, got %T: %v", err, err)
	}

	if exitErr.ExitCode() != exitcodes.General {
		t.Fatalf("exit code = %d, want %d\n%s", exitErr.ExitCode(), exitcodes.General, output)
	}

	want := "cannot clone to " + destinationPath + ": destination exists but is not a git repository. Move or remove it first, then run gcm clone again"
	if !strings.Contains(string(output), want) {
		t.Fatalf("gcm clone output = %q, want actionable error %q", output, want)
	}
}

func TestClonedRepositoryAppearsInStatusOutputImmediately(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")
	remoteURL := "file://localhost" + createBareRemote(t)

	cloneCommand := exec.Command(binary, "clone", remoteURL)
	cloneCommand.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	if output, err := cloneCommand.CombinedOutput(); err != nil {
		t.Fatalf("gcm clone failed: %v\n%s", err, output)
	}

	parts, err := repourl.Parse(remoteURL)
	if err != nil {
		t.Fatalf("parse remote URL for expected path: %v", err)
	}

	destinationPath := derivedpath.Derive(cloneRoot, parts.Hostname, parts.PathPrefix, parts.RepositoryName)
	relativePath, err := filepath.Rel(cloneRoot, destinationPath)
	if err != nil {
		t.Fatalf("compute relative path: %v", err)
	}

	statusCommand := exec.Command(binary, "status")
	statusCommand.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := statusCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm status failed: %v\n%s", err, output)
	}

	if !strings.Contains(string(output), relativePath) {
		t.Fatalf("gcm status output = %q, want repository path %q", output, relativePath)
	}
}

func TestStatusShowsFormattedTableForRepositoriesUnderCloneRoot(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")

	currentRemote := createBareRemote(t)
	currentRepoPath := filepath.Join(cloneRoot, "github.com", "acme", "current")
	cloneRemoteTo(t, currentRemote, currentRepoPath)

	behindRemote := createBareRemote(t)
	behindRepoPath := filepath.Join(cloneRoot, "github.com", "acme", "behind")
	cloneRemoteTo(t, behindRemote, behindRepoPath)
	pushCommitToRemote(t, behindRemote, "second commit")

	featureRemote := createBareRemote(t)
	featureRepoPath := filepath.Join(cloneRoot, "github.com", "acme", "feature")
	cloneRemoteTo(t, featureRemote, featureRepoPath)
	runGit(t, featureRepoPath, "checkout", "-b", "feature/login")
	writeFile(t, filepath.Join(featureRepoPath, "notes.txt"), "draft\n")
	pushCommitToRemote(t, featureRemote, "second commit")

	command := exec.Command(binary, "status")
	command.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm status failed: %v\n%s", err, output)
	}

	status := string(output)
	if !strings.Contains(status, "Repos under "+cloneRoot+":\n") {
		t.Fatalf("gcm status output = %q, want header for clone root %q", status, cloneRoot)
	}

	featureRow := "github.com/acme/feature  feature/login  behind=1  dirty=1  [behind] [!main]"
	behindRow := "github.com/acme/behind   main           behind=1  dirty=0  [behind]"
	currentRow := "github.com/acme/current  main           behind=0  dirty=0"
	for _, want := range []string{
		featureRow,
		behindRow,
		currentRow,
		"3 repos — 1 current, 1 behind, 1 non-default-branch",
		"Tips: gcm pull; gcm status --non-default",
	} {
		if !strings.Contains(status, want) {
			t.Fatalf("gcm status output = %q, want %q", status, want)
		}
	}

	if strings.Index(status, featureRow) > strings.Index(status, behindRow) {
		t.Fatalf("gcm status output = %q, want non-default row before behind row", status)
	}

	if strings.Index(status, behindRow) > strings.Index(status, currentRow) {
		t.Fatalf("gcm status output = %q, want behind row before current row", status)
	}
}

func TestStatusNonDefaultFiltersTable(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")

	currentRemote := createBareRemote(t)
	currentRepoPath := filepath.Join(cloneRoot, "github.com", "acme", "current")
	cloneRemoteTo(t, currentRemote, currentRepoPath)

	featureRemote := createBareRemote(t)
	featureRepoPath := filepath.Join(cloneRoot, "github.com", "acme", "feature")
	cloneRemoteTo(t, featureRemote, featureRepoPath)
	runGit(t, featureRepoPath, "checkout", "-b", "feature/login")

	command := exec.Command(binary, "status", "--non-default")
	command.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm status --non-default failed: %v\n%s", err, output)
	}

	status := string(output)
	if !strings.Contains(status, "github.com/acme/feature  feature/login  behind=0  dirty=0  [!main]") {
		t.Fatalf("gcm status --non-default output = %q, want non-default repository row", status)
	}

	if strings.Contains(status, "github.com/acme/current") {
		t.Fatalf("gcm status --non-default output = %q, did not want default-branch repository row", status)
	}

	if !strings.Contains(status, "1 repos — 0 current, 0 behind, 1 non-default-branch") {
		t.Fatalf("gcm status --non-default output = %q, want filtered summary counts", status)
	}
}

func TestStatusNoFetchUsesLocalStateWhenRemoteIsUnreachable(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")

	remotePath := createBareRemote(t)
	repoPath := filepath.Join(cloneRoot, "github.com", "acme", "current")
	cloneRemoteTo(t, remotePath, repoPath)
	runGit(t, repoPath, "remote", "set-url", "origin", "ssh://127.0.0.1:1/example/repo.git")

	command := exec.Command(binary, "status", "--no-fetch")
	command.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm status --no-fetch failed: %v\n%s", err, output)
	}

	status := string(output)
	if strings.Contains(status, "[fetch-failed]") {
		t.Fatalf("gcm status --no-fetch output = %q, did not want fetch-failed marker", status)
	}

	if !strings.Contains(status, "github.com/acme/current") || !strings.Contains(status, "behind=0  dirty=0") {
		t.Fatalf("gcm status --no-fetch output = %q, want repository row from local state", status)
	}
}

func TestStatusReturnsNonZeroAndShowsFetchFailedRepositories(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")

	remotePath := createBareRemote(t)
	repoPath := filepath.Join(cloneRoot, "github.com", "acme", "offline")
	cloneRemoteTo(t, remotePath, repoPath)
	runGit(t, repoPath, "remote", "set-url", "origin", "ssh://127.0.0.1:1/example/repo.git")

	command := exec.Command(binary, "status")
	command.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("gcm status unexpectedly succeeded:\n%s", output)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exit error, got %T: %v", err, err)
	}

	if exitErr.ExitCode() != exitcodes.General {
		t.Fatalf("exit code = %d, want %d\n%s", exitErr.ExitCode(), exitcodes.General, output)
	}

	status := string(output)
	if !strings.Contains(status, "[fetch-failed]") {
		t.Fatalf("gcm status output = %q, want fetch-failed marker", status)
	}

	if !strings.Contains(status, "one or more repositories failed to fetch") {
		t.Fatalf("gcm status output = %q, want partial-failure message", status)
	}
}

func TestStatusShowsNoRemoteRepositoriesWithoutFailing(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")

	repoPath := filepath.Join(cloneRoot, "github.com", "acme", "local-only")
	if err := os.MkdirAll(filepath.Dir(repoPath), 0o755); err != nil {
		t.Fatalf("create repository parent directory: %v", err)
	}
	runGitInDir(t, "", "init", "--initial-branch=main", repoPath)
	writeFile(t, filepath.Join(repoPath, "README.md"), "hello\n")
	runGit(t, repoPath, "add", "README.md")
	runGit(t, repoPath, "commit", "-m", "initial commit")

	command := exec.Command(binary, "status")
	command.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm status failed: %v\n%s", err, output)
	}

	status := string(output)
	if !strings.Contains(status, "[no-remote]") {
		t.Fatalf("gcm status output = %q, want no-remote marker", status)
	}

	if !strings.Contains(status, "1 repos — 0 current, 0 behind, 0 non-default-branch") {
		t.Fatalf("gcm status output = %q, want no-remote repository excluded from summary counts", status)
	}
}

func TestStatusCompletesWithinTenSecondsForTwoHundredRepositories(t *testing.T) {
	binary := buildGCM(t)
	cloneRoot := filepath.Join(t.TempDir(), "src")
	configPath := writeConfigFile(t, "clone_root: "+cloneRoot+"\n")
	remotePath := createBareRemote(t)

	for index := range 200 {
		repoPath := filepath.Join(cloneRoot, "github.com", "acme", fmt.Sprintf("repo-%03d", index))
		cloneRemoteTo(t, remotePath, repoPath)
	}

	command := exec.Command(binary, "status")
	command.Env = append(os.Environ(), "GCM_CONFIG="+configPath)

	start := time.Now()
	output, err := command.CombinedOutput()
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("gcm status failed: %v\n%s", err, output)
	}

	if duration > 10*time.Second {
		t.Fatalf("gcm status duration = %v, want <= %v", duration, 10*time.Second)
	}

	if !strings.Contains(string(output), "200 repos —") {
		t.Fatalf("gcm status output = %q, want 200-repository summary", output)
	}
}

func TestConfigShowPrintsDefaultCloneRootWithoutCreatingConfigFile(t *testing.T) {
	binary := buildGCM(t)
	configPath := filepath.Join(t.TempDir(), "gcm", "config.yaml")

	cmd := exec.Command(binary, "config", "show")
	cmd.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm config show failed: %v\n%s", err, output)
	}

	if got := string(output); got != "clone_root: ~/src  # default\n" {
		t.Fatalf("gcm config show output = %q, want %q", got, "clone_root: ~/src  # default\n")
	}

	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config file existence = %v, want %v", err, os.ErrNotExist)
	}
}

func TestConfigShowPrintsConfiguredCloneRoot(t *testing.T) {
	binary := buildGCM(t)
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("clone_root: /custom/path\n"), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cmd := exec.Command(binary, "config", "show")
	cmd.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm config show failed: %v\n%s", err, output)
	}

	if got := string(output); got != "clone_root: /custom/path\n" {
		t.Fatalf("gcm config show output = %q, want %q", got, "clone_root: /custom/path\n")
	}
}

func TestConfigSetCloneRootWritesConfigAndPrintsSavedPath(t *testing.T) {
	binary := buildGCM(t)
	configPath := filepath.Join(t.TempDir(), "nested", "gcm", "config.yaml")

	cmd := exec.Command(binary, "config", "set", "clone-root", "/custom/path")
	cmd.Env = append(os.Environ(), "GCM_CONFIG="+configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm config set clone-root failed: %v\n%s", err, output)
	}

	if got := string(output); got != "Config saved to "+configPath+"\n" {
		t.Fatalf("gcm config set clone-root output = %q, want %q", got, "Config saved to "+configPath+"\n")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}

	if got := string(data); got != "clone_root: /custom/path\n" {
		t.Fatalf("config file contents = %q, want %q", got, "clone_root: /custom/path\n")
	}

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}

	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config file mode = %o, want %o", got, 0o600)
	}
}

func TestConfigShowReflectsCloneRootWrittenByConfigSet(t *testing.T) {
	binary := buildGCM(t)
	configPath := filepath.Join(t.TempDir(), "custom", "config.yaml")
	env := append(os.Environ(), "GCM_CONFIG="+configPath)

	setCommand := exec.Command(binary, "config", "set", "clone-root", "/custom/path")
	setCommand.Env = env
	if output, err := setCommand.CombinedOutput(); err != nil {
		t.Fatalf("gcm config set clone-root failed: %v\n%s", err, output)
	}

	showCommand := exec.Command(binary, "config", "show")
	showCommand.Env = env
	output, err := showCommand.CombinedOutput()
	if err != nil {
		t.Fatalf("gcm config show failed: %v\n%s", err, output)
	}

	if got := string(output); got != "clone_root: /custom/path\n" {
		t.Fatalf("gcm config show output = %q, want %q", got, "clone_root: /custom/path\n")
	}
}

func TestUsageErrorsExitWithCode2(t *testing.T) {
	binary := buildGCM(t)

	cmd := exec.Command(binary, "config", "set", "clone-root")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("gcm config set clone-root unexpectedly succeeded:\n%s", output)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exit error, got %T: %v", err, err)
	}

	if exitErr.ExitCode() != exitcodes.Usage {
		t.Fatalf("exit code = %d, want %d\n%s", exitErr.ExitCode(), exitcodes.Usage, output)
	}
}

func TestCloneMissingURLArgumentExitsWithCode2(t *testing.T) {
	binary := buildGCM(t)

	cmd := exec.Command(binary, "clone")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("gcm clone unexpectedly succeeded:\n%s", output)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exit error, got %T: %v", err, err)
	}

	if exitErr.ExitCode() != exitcodes.Usage {
		t.Fatalf("exit code = %d, want %d\n%s", exitErr.ExitCode(), exitcodes.Usage, output)
	}
}

func buildGCM(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	binary := filepath.Join(tempDir, "gcm")

	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	return binary
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

func cloneRemoteTo(t *testing.T, remotePath, destinationPath string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		t.Fatalf("create clone parent directory: %v", err)
	}

	runGitInDir(t, "", "clone", remotePath, destinationPath)
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

func writeConfigFile(t *testing.T, contents string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "gcm", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	return configPath
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
