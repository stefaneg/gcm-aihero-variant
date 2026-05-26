package gitrunner

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Runner interface {
	Clone(url, destPath string) error
	OriginURL(repoPath string) (string, error)
	Fetch(repoPath string) error
	CurrentBranch(repoPath string) (string, error)
	DirtyCount(repoPath string) (int, error)
	CommitsBehind(repoPath string) (int, error)
	DefaultBranch(repoPath string) (string, error)
}

type GitNotFoundError struct {
	Binary string
	Err    error
}

func (errorWithContext *GitNotFoundError) Error() string {
	return fmt.Sprintf("git binary %q not found: %v", errorWithContext.Binary, errorWithContext.Err)
}

func (errorWithContext *GitNotFoundError) Unwrap() error {
	return errorWithContext.Err
}

type RepositoryNotFoundError struct {
	RepositoryPath string
	Err            error
}

func (errorWithContext *RepositoryNotFoundError) Error() string {
	return fmt.Sprintf("repository %q not found: %v", errorWithContext.RepositoryPath, errorWithContext.Err)
}

func (errorWithContext *RepositoryNotFoundError) Unwrap() error {
	return errorWithContext.Err
}

type NetworkError struct {
	Operation      string
	RepositoryPath string
	Err            error
}

func (errorWithContext *NetworkError) Error() string {
	return fmt.Sprintf("git %s failed for %q: %v", errorWithContext.Operation, errorWithContext.RepositoryPath, errorWithContext.Err)
}

func (errorWithContext *NetworkError) Unwrap() error {
	return errorWithContext.Err
}

type NoRemoteError struct {
	RepositoryPath string
	Err            error
}

func (errorWithContext *NoRemoteError) Error() string {
	return fmt.Sprintf("repository %q has no remote configured: %v", errorWithContext.RepositoryPath, errorWithContext.Err)
}

func (errorWithContext *NoRemoteError) Unwrap() error {
	return errorWithContext.Err
}

type OriginHEADNotSetError struct {
	RepositoryPath string
	Err            error
}

func (errorWithContext *OriginHEADNotSetError) Error() string {
	return fmt.Sprintf("repository %q does not have refs/remotes/origin/HEAD set: %v", errorWithContext.RepositoryPath, errorWithContext.Err)
}

func (errorWithContext *OriginHEADNotSetError) Unwrap() error {
	return errorWithContext.Err
}

type runner struct {
	gitBinary string
}

func New() Runner {
	return &runner{gitBinary: "git"}
}

func NewForTesting(gitBinary string) Runner {
	return &runner{gitBinary: gitBinary}
}

func (gitRunner *runner) Clone(url, destPath string) error {
	_, err := gitRunner.run("", "clone", url, destPath)
	if err != nil {
		return gitRunner.classifyError("clone", destPath, err)
	}

	return nil
}

func (gitRunner *runner) OriginURL(repoPath string) (string, error) {
	output, err := gitRunner.run(repoPath, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", gitRunner.classifyError("config", repoPath, err)
	}

	return strings.TrimSpace(output), nil
}

func (gitRunner *runner) Fetch(repoPath string) error {
	_, err := gitRunner.run(repoPath, "fetch", "origin")
	if err != nil {
		return gitRunner.classifyError("fetch", repoPath, err)
	}

	return nil
}

func (gitRunner *runner) DirtyCount(repoPath string) (int, error) {
	output, err := gitRunner.run(repoPath, "status", "--porcelain")
	if err != nil {
		return 0, gitRunner.classifyError("status", repoPath, err)
	}

	return countNonEmptyLines(output), nil
}

func (gitRunner *runner) CurrentBranch(repoPath string) (string, error) {
	output, err := gitRunner.run(repoPath, "branch", "--show-current")
	if err != nil {
		return "", gitRunner.classifyError("branch", repoPath, err)
	}

	return strings.TrimSpace(output), nil
}

func (gitRunner *runner) CommitsBehind(repoPath string) (int, error) {
	defaultBranch, err := gitRunner.DefaultBranch(repoPath)
	if err != nil {
		return 0, err
	}

	output, err := gitRunner.run(
		repoPath,
		"rev-list",
		"--count",
		fmt.Sprintf("refs/heads/%s..refs/remotes/origin/%s", defaultBranch, defaultBranch),
	)
	if err != nil {
		return 0, gitRunner.classifyError("rev-list", repoPath, err)
	}

	commitsBehind, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("parse commits behind count: %w", err)
	}

	return commitsBehind, nil
}

func (gitRunner *runner) DefaultBranch(repoPath string) (string, error) {
	output, err := gitRunner.run(repoPath, "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", gitRunner.classifyError("symbolic-ref", repoPath, err)
	}

	defaultBranchRef := strings.TrimSpace(output)
	if !strings.HasPrefix(defaultBranchRef, "origin/") {
		return "", fmt.Errorf("origin HEAD ref %q does not point at origin/<branch>", defaultBranchRef)
	}

	return strings.TrimPrefix(defaultBranchRef, "origin/"), nil
}

func (gitRunner *runner) run(repoPath string, args ...string) (string, error) {
	commandArgs := args
	if repoPath != "" {
		commandArgs = append([]string{"-C", repoPath}, args...)
	}

	cmd := exec.Command(gitRunner.gitBinary, commandArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		if stderr.Len() > 0 {
			return "", &commandError{stderr: strings.TrimSpace(stderr.String()), err: err}
		}

		return "", err
	}

	return string(output), nil
}

func (gitRunner *runner) classifyError(operation, repoPath string, err error) error {
	if errors.Is(err, exec.ErrNotFound) {
		return &GitNotFoundError{Binary: gitRunner.gitBinary, Err: err}
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		var commandErr *commandError
		if errors.As(err, &commandErr) {
			stderr := strings.ToLower(commandErr.stderr)
			switch {
			case strings.Contains(stderr, "cannot change to"):
				return &RepositoryNotFoundError{RepositoryPath: repoPath, Err: err}
			case strings.Contains(stderr, "not a git repository"):
				return &RepositoryNotFoundError{RepositoryPath: repoPath, Err: err}
			case strings.Contains(stderr, "no such file or directory"):
				return &RepositoryNotFoundError{RepositoryPath: repoPath, Err: err}
			case operation == "symbolic-ref" && strings.Contains(stderr, "is not a symbolic ref"):
				return &OriginHEADNotSetError{RepositoryPath: repoPath, Err: err}
			case strings.Contains(stderr, "no remote repository specified"),
				strings.Contains(stderr, "no configured push destination"),
				strings.Contains(stderr, "does not appear to be a git repository"),
				strings.Contains(stderr, "no upstream configured"):
				return &NoRemoteError{RepositoryPath: repoPath, Err: err}
			case strings.Contains(stderr, "could not resolve host"),
				strings.Contains(stderr, "connection refused"),
				strings.Contains(stderr, "operation timed out"),
				strings.Contains(stderr, "network is unreachable"),
				strings.Contains(stderr, "could not read from remote repository"),
				strings.Contains(stderr, "unable to access"):
				return &NetworkError{Operation: operation, RepositoryPath: repoPath, Err: err}
			}
		}
	}

	return err
}

type commandError struct {
	stderr string
	err    error
}

func (errorWithContext *commandError) Error() string {
	if errorWithContext.stderr == "" {
		return errorWithContext.err.Error()
	}

	return errorWithContext.stderr
}

func (errorWithContext *commandError) Unwrap() error {
	return errorWithContext.err
}

func countNonEmptyLines(output string) int {
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count
}
