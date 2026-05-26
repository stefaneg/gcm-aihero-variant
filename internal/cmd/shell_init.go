package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"git-clone-manager/internal/exitcodes"

	"github.com/spf13/cobra"
)

func newShellInitCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "shell-init [bash|zsh|fish]",
		Short: "Print shell integration for changing directory after clone",
		RunE: func(command *cobra.Command, args []string) error {
			install, err := command.Flags().GetBool("install")
			if err != nil {
				return err
			}

			shell, err := resolveShell(args)
			if err != nil {
				return err
			}

			if install {
				return installShellInit(shell, command.ErrOrStderr())
			}

			if _, err := fmt.Fprint(command.OutOrStdout(), shellWrapper(shell)); err != nil {
				return err
			}

			if writerIsTTY(command.OutOrStdout()) {
				if _, err := fmt.Fprint(command.ErrOrStderr(), installHint(shell)); err != nil {
					return err
				}
			}

			return nil
		},
	}
	command.Args = usageArgs(cobra.MaximumNArgs(1))
	command.Flags().Bool("install", false, "Install shell integration in the detected shell rc file")
	return command
}

func resolveShell(args []string) (string, error) {
	if len(args) > 0 {
		return validateShell(args[0])
	}

	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return "", exitcodes.UsageError(fmt.Errorf(`Error: environment variable "SHELL": unset`))
	}

	return validateShell(filepath.Base(shellPath))
}

func validateShell(shell string) (string, error) {
	switch shell {
	case "bash", "zsh", "fish":
		return shell, nil
	default:
		return "", exitcodes.UsageError(fmt.Errorf(`Error: shell %q: unsupported shell`, shell))
	}
}

func shellWrapper(shell string) string {
	switch shell {
	case "fish":
		return `function gcm
    if test (count $argv) -gt 0; and test "$argv[1]" = "clone"
        set -l dest (command gcm $argv)
        set -l command_status $status
        if test $command_status -eq 0; and test -n "$dest"
            cd "$dest"
        end
        return $command_status
    else
        command gcm $argv
    end
end
`
	default:
		return `gcm() {
  if [ "$1" = "clone" ]; then
    local dest
    dest=$(command gcm "$@")
    if [ $status -eq 0 ] && [ -n "$dest" ]; then
      cd "$dest"
    fi
    return $status
  else
    command gcm "$@"
  fi
}
`
	}
}

func installHint(shell string) string {
	return fmt.Sprintf("To install permanently, add this line to %s:\n  %s\nOr run: gcm shell-init --install\n", rcDisplayPath(shell), installLine(shell))
}

func installShellInit(shell string, stderr io.Writer) error {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return exitcodes.UsageError(fmt.Errorf(`Error: environment variable "HOME": unset`))
	}

	rcPath, err := rcFilePath(shell, homeDir)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(rcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error: rc file %q: %w", rcPath, err)
	}

	if strings.Contains(string(data), "gcm shell-init") {
		_, err := fmt.Fprintf(stderr, "Already installed in %s.\n", rcPath)
		return err
	}

	if err := os.MkdirAll(filepath.Dir(rcPath), 0o755); err != nil {
		return fmt.Errorf("Error: rc file %q: %w", rcPath, err)
	}

	file, err := os.OpenFile(rcPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("Error: rc file %q: %w", rcPath, err)
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "\n# gcm shell-init\n%s\n", installLine(shell)); err != nil {
		return fmt.Errorf("Error: rc file %q: %w", rcPath, err)
	}

	_, err = fmt.Fprintf(stderr, "Installed in %s. Reload your shell or run: source %s\n", rcPath, rcPath)
	return err
}

func rcFilePath(shell string, homeDir string) (string, error) {
	switch shell {
	case "bash":
		bashRC := filepath.Join(homeDir, ".bashrc")
		bashProfile := filepath.Join(homeDir, ".bash_profile")
		if _, err := os.Stat(bashRC); err == nil {
			return bashRC, nil
		}
		if _, err := os.Stat(bashProfile); err == nil {
			return bashProfile, nil
		}
		return bashRC, nil
	case "zsh":
		return filepath.Join(homeDir, ".zshrc"), nil
	case "fish":
		return filepath.Join(homeDir, ".config", "fish", "config.fish"), nil
	default:
		return "", exitcodes.UsageError(fmt.Errorf(`Error: shell %q: unsupported shell`, shell))
	}
}

func rcDisplayPath(shell string) string {
	switch shell {
	case "bash":
		return "~/.bashrc"
	case "zsh":
		return "~/.zshrc"
	case "fish":
		return "~/.config/fish/config.fish"
	default:
		return "~/.profile"
	}
}

func installLine(shell string) string {
	if shell == "fish" {
		return "gcm shell-init fish | source"
	}

	return `eval "$(gcm shell-init)"`
}
