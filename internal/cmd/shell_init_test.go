package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellInitPrintsWrapperForRequestedShell(t *testing.T) {
	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))

	tests := []struct {
		name string
		arg  string
		want []string
	}{
		{
			name: "bash",
			arg:  "bash",
			want: []string{"gcm() {", `if [ "$1" = "clone" ]; then`, `if [ "$1" = "open" ]; then`, `command gcm "$@"`, `cd "$dest"`},
		},
		{
			name: "zsh",
			arg:  "zsh",
			want: []string{"gcm() {", `if [ "$1" = "clone" ]; then`, `if [ "$1" = "open" ]; then`, `command gcm "$@"`, `cd "$dest"`},
		},
		{
			name: "fish",
			arg:  "fish",
			want: []string{"function gcm", `test "$argv[1]" = "clone"`, `test "$argv[1]" = "open"`, "command gcm $argv", `cd "$dest"`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			exitCode := Execute([]string{"shell-init", test.arg}, &stdout, &stderr)
			if exitCode != 0 {
				t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
			}

			for _, want := range test.want {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout = %q, want %q", stdout.String(), want)
				}
			}

			if stderr.String() != "" {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestShellInitDetectsShellFromEnvironment(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"shell-init"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), "gcm() {") {
		t.Fatalf("stdout = %q, want zsh wrapper", stdout.String())
	}

	if strings.Contains(stdout.String(), "$status") {
		t.Fatalf("stdout = %q, did not expect reserved zsh status variable usage", stdout.String())
	}

	if !strings.Contains(stdout.String(), "local command_status=$?") {
		t.Fatalf("stdout = %q, want explicit command status capture", stdout.String())
	}
}

func TestShellInitBakesProjectOpenerIntoPosixWrapper(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("GCM_CONFIG", configPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Execute([]string{"config", "set", "project-opener", "code --new-window"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("config set exit code = %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	exitCode := Execute([]string{"shell-init", "zsh"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("shell-init exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), `cd "$dest" && code --new-window .`) {
		t.Fatalf("stdout = %q, want baked project opener", stdout.String())
	}
	if strings.Contains(stdout.String(), "gcm config") || strings.Contains(stdout.String(), "GCM_OPENER") || strings.Contains(stdout.String(), "EDITOR") || strings.Contains(stdout.String(), "VISUAL") {
		t.Fatalf("stdout = %q, wrapper must not resolve opener at runtime", stdout.String())
	}
}

func TestShellInitBakesProjectOpenerIntoFishWrapper(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("GCM_CONFIG", configPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Execute([]string{"config", "set", "project-opener", "goland"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("config set exit code = %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	exitCode := Execute([]string{"shell-init", "fish"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("shell-init exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), `cd "$dest"; and goland .`) {
		t.Fatalf("stdout = %q, want baked fish project opener", stdout.String())
	}
	if strings.Contains(stdout.String(), "gcm config") || strings.Contains(stdout.String(), "GCM_OPENER") || strings.Contains(stdout.String(), "EDITOR") || strings.Contains(stdout.String(), "VISUAL") {
		t.Fatalf("stdout = %q, wrapper must not resolve opener at runtime", stdout.String())
	}
}

func TestShellInitWithoutProjectOpenerKeepsOpenBranchCdOnly(t *testing.T) {
	t.Setenv("GCM_CONFIG", filepath.Join(t.TempDir(), "config.yaml"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"shell-init", "zsh"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("shell-init exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), `cd "$dest"`) {
		t.Fatalf("stdout = %q, want cd-only open branch", stdout.String())
	}
	if strings.Contains(stdout.String(), `cd "$dest" &&`) {
		t.Fatalf("stdout = %q, did not expect opener launch", stdout.String())
	}
}

func TestShellInitRejectsUnsupportedShell(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"shell-init", "tcsh"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("Execute exit code = %d, want 2\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	if !strings.Contains(stderr.String(), `"tcsh"`) {
		t.Fatalf("stderr = %q, want unsupported shell named literally", stderr.String())
	}
}

func TestShellInitInstallWritesAndReusesZshRC(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("SHELL", "/bin/zsh")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"shell-init", "--install"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	rcPath := filepath.Join(homeDir, ".zshrc")
	data, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("read .zshrc: %v", err)
	}

	if !strings.Contains(string(data), `eval "$(gcm shell-init)"`) {
		t.Fatalf(".zshrc = %q, want install line", data)
	}
	if !strings.Contains(string(data), shellInitInstallMarker) {
		t.Fatalf(".zshrc = %q, want install marker", data)
	}

	firstContents := string(data)
	stderr.Reset()

	exitCode = Execute([]string{"shell-init", "--install"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("second Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	data, err = os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("read .zshrc after second install: %v", err)
	}

	if string(data) != firstContents {
		t.Fatalf(".zshrc changed on second install:\nfirst %q\nsecond %q", firstContents, data)
	}

	if !strings.Contains(stderr.String(), "Already installed in "+rcPath+".") {
		t.Fatalf("stderr = %q, want already-installed message", stderr.String())
	}
}

func TestShellInitInstallIgnoresCommentMentioningShellInit(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("SHELL", "/bin/zsh")
	rcPath := filepath.Join(homeDir, ".zshrc")
	if err := os.WriteFile(rcPath, []byte("# disabled: gcm shell-init\n"), 0o644); err != nil {
		t.Fatalf("write .zshrc: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"shell-init", "--install"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	data, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("read .zshrc: %v", err)
	}
	if !strings.Contains(string(data), shellInitInstallMarker) || !strings.Contains(string(data), `eval "$(gcm shell-init)"`) {
		t.Fatalf(".zshrc = %q, want marker and install line", data)
	}
}

func TestShellInitInstallKeepsLegacyInstallIdempotent(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("SHELL", "/bin/zsh")
	rcPath := filepath.Join(homeDir, ".zshrc")
	legacy := "\n# gcm shell-init\n" + installLine("zsh") + "\n"
	if err := os.WriteFile(rcPath, []byte(legacy), 0o644); err != nil {
		t.Fatalf("write .zshrc: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"shell-init", "--install"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	data, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("read .zshrc: %v", err)
	}
	if string(data) != legacy {
		t.Fatalf(".zshrc changed for legacy install:\n%s", data)
	}
	if !strings.Contains(stderr.String(), "Already installed in "+rcPath+".") {
		t.Fatalf("stderr = %q, want already-installed message", stderr.String())
	}
}
