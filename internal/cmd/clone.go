package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"git-clone-manager/internal/gitrunner"
	"git-clone-manager/internal/repourl"

	"github.com/spf13/cobra"
)

func newCloneCommand(deps Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "clone <url>",
		Short: "Clone a repository into its derived path",
		RunE: func(command *cobra.Command, args []string) error {
			effectiveConfig, err := deps.LoadEffectiveCloneConfig()
			if err != nil {
				return err
			}

			cloneRoot, err := expandHomePath(effectiveConfig.CloneRoot)
			if err != nil {
				return err
			}

			parts, err := repourl.Parse(args[0])
			if err != nil {
				return err
			}

			destinationPath := parts.DerivedPath(cloneRoot)
			if err := ensureCloneRoot(command.ErrOrStderr(), cloneRoot); err != nil {
				return err
			}

			runner := deps.NewGitRunner()
			state, err := inspectDestination(destinationPath, args[0], runner)
			if err != nil {
				return err
			}

			switch state {
			case destinationAlreadyCloned:
				_, err := fmt.Fprintln(command.OutOrStdout(), destinationPath)
				return err
			case destinationBlocked:
				return fmt.Errorf("cannot clone to %s: destination exists but is not a git repository. Move or remove it first, then run gcm clone again", destinationPath)
			}

			preExistingDestination := state == destinationReadyPreExisting
			createdDirs, err := mkdirAllTracked(filepath.Dir(destinationPath), 0o755)
			if err != nil {
				return fmt.Errorf("create parent directories for %q: %w", destinationPath, err)
			}

			if _, err := fmt.Fprintln(command.ErrOrStderr(), "Cloning to "+destinationPath+"..."); err != nil {
				return err
			}

			if err := runner.Clone(context.Background(), args[0], destinationPath); err != nil {
				cleanupPartialClone(destinationPath, preExistingDestination, createdDirs)
				return err
			}

			if _, err := fmt.Fprintln(command.ErrOrStderr(), "Done."); err != nil {
				return err
			}

			_, err = fmt.Fprintln(command.OutOrStdout(), destinationPath)
			return err
		},
	}
	command.Args = usageArgs(cobra.ExactArgs(1))
	return command
}

type destinationState int

const (
	destinationReady destinationState = iota
	destinationReadyPreExisting
	destinationAlreadyCloned
	destinationBlocked
)

func expandHomePath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("determine home directory for clone root %q: %w", path, err)
		}

		if path == "~" {
			return homeDir, nil
		}

		return filepath.Join(homeDir, strings.TrimPrefix(path, "~/")), nil
	}

	return path, nil
}

func ensureCloneRoot(stderr io.Writer, cloneRoot string) error {
	if info, err := os.Stat(cloneRoot); err == nil {
		if info.IsDir() {
			return nil
		}

		return fmt.Errorf("cannot use clone root %s: path exists but is not a directory. Move or remove it, then run gcm clone again", cloneRoot)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat clone root %q: %w", cloneRoot, err)
	}

	if _, err := fmt.Fprintln(stderr, "Clone root "+cloneRoot+" does not exist - creating it"); err != nil {
		return err
	}

	if err := os.MkdirAll(cloneRoot, 0o755); err != nil {
		return fmt.Errorf("create clone root %q: %w", cloneRoot, err)
	}

	return nil
}

func inspectDestination(destinationPath string, requestedURL string, runner gitrunner.Runner) (destinationState, error) {
	info, err := os.Stat(destinationPath)
	if errors.Is(err, os.ErrNotExist) {
		return destinationReady, nil
	}
	if err != nil {
		return destinationReady, fmt.Errorf("stat destination %q: %w", destinationPath, err)
	}

	if !info.IsDir() {
		return destinationBlocked, nil
	}

	gitDirInfo, err := os.Stat(filepath.Join(destinationPath, ".git"))
	if err == nil && gitDirInfo.IsDir() {
		originURL, err := runner.OriginURL(context.Background(), destinationPath)
		if err != nil {
			return destinationBlocked, nil
		}
		if originURL != requestedURL {
			return destinationReady, fmt.Errorf("cannot clone to %s: existing git repository has origin %q, not requested URL %q", destinationPath, originURL, requestedURL)
		}
		return destinationAlreadyCloned, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return destinationReady, fmt.Errorf("inspect destination %q: %w", destinationPath, err)
	}

	entries, err := os.ReadDir(destinationPath)
	if err != nil {
		return destinationReady, fmt.Errorf("inspect destination %q: %w", destinationPath, err)
	}
	if len(entries) == 0 {
		return destinationReadyPreExisting, nil
	}

	return destinationBlocked, nil
}

func mkdirAllTracked(path string, perm os.FileMode) ([]string, error) {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || cleanPath == string(filepath.Separator) {
		return nil, nil
	}

	var segments []string
	for current := cleanPath; current != "." && current != string(filepath.Separator); current = filepath.Dir(current) {
		segments = append(segments, current)
	}
	slices.Reverse(segments)

	var created []string
	for _, segment := range segments {
		if err := os.Mkdir(segment, perm); err == nil {
			created = append(created, segment)
			continue
		} else if errors.Is(err, os.ErrExist) {
			info, statErr := os.Stat(segment)
			if statErr != nil {
				return nil, statErr
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("create directory %q: path exists but is not a directory", segment)
			}
			continue
		} else {
			return nil, err
		}
	}

	slices.Reverse(created)
	return created, nil
}

func cleanupPartialClone(destinationPath string, preExistingDestination bool, createdDirs []string) {
	if preExistingDestination {
		entries, err := os.ReadDir(destinationPath)
		if err == nil {
			for _, entry := range entries {
				_ = os.RemoveAll(filepath.Join(destinationPath, entry.Name()))
			}
		}
	} else {
		_ = os.RemoveAll(destinationPath)
	}

	for _, dir := range createdDirs {
		_ = os.Remove(dir)
	}
}
