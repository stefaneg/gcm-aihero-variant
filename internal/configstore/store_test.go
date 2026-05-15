package configstore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEffectiveReturnsDefaultCloneRootWithoutCreatingConfigFile(t *testing.T) {
	tempHome := t.TempDir()
	store := &Store{
		lookupEnv: func(string) string { return "" },
		userHomeDir: func() (string, error) {
			return tempHome, nil
		},
	}

	effectiveConfig, err := store.Effective()
	if err != nil {
		t.Fatalf("Effective() error = %v", err)
	}

	if effectiveConfig.CloneRoot != defaultCloneRoot {
		t.Fatalf("Effective().CloneRoot = %q, want %q", effectiveConfig.CloneRoot, defaultCloneRoot)
	}

	if !effectiveConfig.CloneRootIsDefault {
		t.Fatalf("Effective().CloneRootIsDefault = %t, want %t", effectiveConfig.CloneRootIsDefault, true)
	}

	configPath := filepath.Join(tempHome, ".config", "gcm", defaultConfigName)
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("config file existence = %v, want %v", err, os.ErrNotExist)
	}
}

func TestPathUsesGCMConfigOverride(t *testing.T) {
	overridePath := filepath.Join(t.TempDir(), "override", "config.yaml")
	store := &Store{
		lookupEnv: func(name string) string {
			if name == configPathEnv {
				return overridePath
			}
			return ""
		},
		userHomeDir: func() (string, error) {
			return t.TempDir(), nil
		},
	}

	configPath, err := store.Path()
	if err != nil {
		t.Fatalf("Path() error = %v", err)
	}

	if configPath != overridePath {
		t.Fatalf("Path() = %q, want %q", configPath, overridePath)
	}
}

func TestSetCloneRootWritesConfigUnderTempHome(t *testing.T) {
	tempHome := t.TempDir()
	store := &Store{
		lookupEnv: func(string) string { return "" },
		userHomeDir: func() (string, error) {
			return tempHome, nil
		},
	}

	configPath, err := store.SetCloneRoot("/custom/path")
	if err != nil {
		t.Fatalf("SetCloneRoot() error = %v", err)
	}

	wantPath := filepath.Join(tempHome, ".config", "gcm", defaultConfigName)
	if configPath != wantPath {
		t.Fatalf("SetCloneRoot() path = %q, want %q", configPath, wantPath)
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
