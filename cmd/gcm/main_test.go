package main_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"git-clone-manager/internal/exitcodes"
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

func TestLeafCommandsStubOutSuccessfully(t *testing.T) {
	binary := buildGCM(t)

	tests := []struct {
		name string
		args []string
	}{
		{name: "clone", args: []string{"clone", "https://github.com/example/repo.git"}},
		{name: "status", args: []string{"status"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := exec.Command(binary, test.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s failed: %v\n%s", strings.Join(test.args, " "), err, output)
			}

			if string(output) != "not yet implemented\n" {
				t.Fatalf("%s output = %q, want %q", strings.Join(test.args, " "), output, "not yet implemented\n")
			}
		})
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
