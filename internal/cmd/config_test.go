package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigShowPrintsDefaultCloneRootWithoutCreatingConfigFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "gcm", "config.yaml")
	t.Setenv("GCM_CONFIG", configPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"config", "show"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if got := stdout.String(); got != "clone_root: ~/src  # default\n" {
		t.Fatalf("stdout = %q, want default clone root", got)
	}

	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config file existence = %v, want %v", err, os.ErrNotExist)
	}
}

func TestConfigSetCloneRootWritesConfigAndPrintsSavedPath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nested", "gcm", "config.yaml")
	t.Setenv("GCM_CONFIG", configPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"config", "set", "clone-root", "/custom/path"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if got := stdout.String(); got != "Config saved to "+configPath+"\n" {
		t.Fatalf("stdout = %q, want saved path", got)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}

	if got := string(data); got != "clone_root: /custom/path\n" {
		t.Fatalf("config file contents = %q, want custom clone root", got)
	}
}

func TestConfigShowReflectsCloneRootWrittenByConfigSet(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "custom", "config.yaml")
	t.Setenv("GCM_CONFIG", configPath)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Execute([]string{"config", "set", "clone-root", "/custom/path"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("config set exit code = %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	if exitCode := Execute([]string{"config", "show"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("config show exit code = %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if got := stdout.String(); got != "clone_root: /custom/path\n" {
		t.Fatalf("stdout = %q, want configured clone root", got)
	}
}
