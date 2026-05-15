package configstore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	configPathEnv     = "GCM_CONFIG"
	defaultCloneRoot  = "~/src"
	defaultConfigName = "config.yaml"
)

type Config struct {
	CloneRoot string
}

type EffectiveConfig struct {
	CloneRoot          string
	CloneRootIsDefault bool
}

type Store struct {
	lookupEnv   func(string) string
	userHomeDir func() (string, error)
}

func New() *Store {
	return &Store{
		lookupEnv:   os.Getenv,
		userHomeDir: os.UserHomeDir,
	}
}

func (store *Store) Path() (string, error) {
	if configuredPath := store.lookupEnv(configPathEnv); configuredPath != "" {
		return configuredPath, nil
	}

	homeDir, err := store.userHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine config path: %w", err)
	}

	return filepath.Join(homeDir, ".config", "gcm", defaultConfigName), nil
}

func (store *Store) Effective() (EffectiveConfig, error) {
	configPath, err := store.Path()
	if err != nil {
		return EffectiveConfig{}, err
	}

	fileConfig, err := store.read(configPath)
	if err != nil {
		return EffectiveConfig{}, err
	}

	if fileConfig.CloneRoot == "" {
		return EffectiveConfig{
			CloneRoot:          defaultCloneRoot,
			CloneRootIsDefault: true,
		}, nil
	}

	return EffectiveConfig{
		CloneRoot:          fileConfig.CloneRoot,
		CloneRootIsDefault: false,
	}, nil
}

func (store *Store) SetCloneRoot(cloneRoot string) (string, error) {
	configPath, err := store.Path()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return "", fmt.Errorf("create config directory for %q: %w", configPath, err)
	}

	if err := os.WriteFile(configPath, []byte(format(Config{CloneRoot: cloneRoot})), 0o600); err != nil {
		return "", fmt.Errorf("write config file %q: %w", configPath, err)
	}

	if err := os.Chmod(configPath, 0o600); err != nil {
		return "", fmt.Errorf("set config file permissions for %q: %w", configPath, err)
	}

	return configPath, nil
}

func (store *Store) read(configPath string) (Config, error) {
	data, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", configPath, err)
	}

	config, err := parse(data)
	if err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", configPath, err)
	}

	return config, nil
}

func parse(data []byte) (Config, error) {
	var config Config

	for lineNumber, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			return Config{}, fmt.Errorf("line %d: missing ':'", lineNumber+1)
		}

		switch strings.TrimSpace(key) {
		case "clone_root":
			parsedValue, err := parseScalar(value)
			if err != nil {
				return Config{}, fmt.Errorf("line %d: %w", lineNumber+1, err)
			}
			config.CloneRoot = parsedValue
		}
	}

	return config, nil
}

func parseScalar(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}

	if strings.HasPrefix(trimmed, "\"") {
		unquoted, err := strconv.Unquote(trimmed)
		if err != nil {
			return "", fmt.Errorf("invalid quoted value: %w", err)
		}
		return unquoted, nil
	}

	return trimmed, nil
}

func format(config Config) string {
	return "clone_root: " + formatScalar(config.CloneRoot) + "\n"
}

func formatScalar(value string) string {
	if value == "" {
		return `""`
	}

	if strings.ContainsAny(value, ":\n#\"'") || strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ") {
		return strconv.Quote(value)
	}

	return value
}
