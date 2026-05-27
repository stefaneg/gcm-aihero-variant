package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git-clone-manager/internal/repositorywalker"

	"github.com/spf13/cobra"
)

func newOpenCommand(deps Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "open [query]",
		Short: "Select a repository under the clone root",
		RunE: func(command *cobra.Command, args []string) error {
			effectiveConfig, err := deps.LoadEffectiveOpenConfig()
			if err != nil {
				return err
			}

			cloneRoot, err := expandHomePath(effectiveConfig.CloneRoot)
			if err != nil {
				return err
			}

			if info, err := os.Stat(cloneRoot); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("clone root %s does not exist. Set it with: gcm config set clone-root <path>", cloneRoot)
				}
				return fmt.Errorf("stat clone root %q: %w", cloneRoot, err)
			} else if !info.IsDir() {
				return fmt.Errorf("clone root %s exists but is not a directory. Set a directory with: gcm config set clone-root <path>", cloneRoot)
			}

			repositoryPaths, err := repositorywalker.Walk(cloneRoot)
			if err != nil {
				return err
			}
			if len(repositoryPaths) == 0 {
				return fmt.Errorf("clone root %s contains no repositories. Clone one with: gcm clone <url>", cloneRoot)
			}

			if !deps.OpenFZFAvailable() {
				return fmt.Errorf("fzf is required for gcm open. Install it from https://github.com/junegunn/fzf#installation")
			}

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			selection, exitCode, err := deps.RunOpenFZF(repositoryPaths, query, openPreviewArgument())
			if err != nil {
				return err
			}
			if exitCode != 0 {
				return nil
			}

			selection = strings.TrimRight(selection, "\r\n")
			if selection == "" {
				return nil
			}

			if _, err := fmt.Fprintln(command.OutOrStdout(), selection); err != nil {
				return err
			}

			if deps.OpenWriterIsTTY(command.OutOrStdout()) {
				_, err := fmt.Fprintln(command.ErrOrStderr(), `gcm open prints a path unless shell integration is installed. Run: eval "$(gcm shell-init)"`)
				return err
			}

			return nil
		},
	}
	command.Args = usageArgs(cobra.MaximumNArgs(1))
	return command
}

func runFZF(repositoryPaths []string, query string, preview string) (string, int, error) {
	args := []string{"--preview", preview, "--preview-window", "right:60%"}
	if query != "" {
		args = append(args, "--query", query)
	}

	command := exec.Command("fzf", args...)
	command.Stdin = strings.NewReader(strings.Join(repositoryPaths, "\n") + "\n")
	command.Stderr = os.Stderr

	var stdout bytes.Buffer
	command.Stdout = &stdout

	err := command.Run()
	if err == nil {
		return stdout.String(), 0, nil
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return stdout.String(), exitError.ExitCode(), nil
	}

	return "", 0, err
}

func openPreviewArgument() string {
	readmeCandidates := []string{
		"README.md",
		"README.markdown",
		"README.txt",
		"README",
		"readme.md",
		"Readme.md",
	}

	var builder strings.Builder
	builder.WriteString(`dir={}; `)
	for _, candidate := range readmeCandidates {
		builder.WriteString(`if [ -f "$dir/`)
		builder.WriteString(candidate)
		builder.WriteString(`" ]; then if command -v bat >/dev/null 2>&1; then bat --style=plain --color=always "$dir/`)
		builder.WriteString(candidate)
		builder.WriteString(`"; else head -n 40 "$dir/`)
		builder.WriteString(candidate)
		builder.WriteString(`"; fi; exit; fi; `)
	}
	builder.WriteString(`echo "No README found in $dir"`)
	return builder.String()
}
