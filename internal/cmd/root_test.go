package cmd

import (
	"bytes"
	"strings"
	"testing"

	"git-clone-manager/internal/exitcodes"
)

func TestHelpListsTopLevelCommands(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Execute([]string{"--help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	help := stdout.String()
	for _, want := range []string{"clone", "status", "config", "shell-init"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help missing %q:\n%s", want, help)
		}
	}
}

func TestCommandHelpRendersUsage(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "clone help", args: []string{"clone", "--help"}, want: []string{"Usage:", "clone <url>"}},
		{name: "status help", args: []string{"status", "--help"}, want: []string{"Usage:", "--no-fetch", "--non-default"}},
		{name: "config help", args: []string{"config", "--help"}, want: []string{"Usage:", "set", "show"}},
		{name: "config set help", args: []string{"config", "set", "--help"}, want: []string{"Usage:", "clone-root"}},
		{name: "config show help", args: []string{"config", "show", "--help"}, want: []string{"Usage:", "show"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			exitCode := Execute(test.args, &stdout, &stderr)
			if exitCode != 0 {
				t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
			}

			help := stdout.String()
			for _, want := range test.want {
				if !strings.Contains(help, want) {
					t.Fatalf("help for %v missing %q:\n%s", test.args, want, help)
				}
			}
		})
	}
}

func TestUsageErrorsExitWithCode2(t *testing.T) {
	tests := [][]string{
		{"config", "set", "clone-root"},
		{"clone"},
	}

	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			exitCode := Execute(args, &stdout, &stderr)
			if exitCode != exitcodes.Usage {
				t.Fatalf("Execute exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", exitCode, exitcodes.Usage, stdout.String(), stderr.String())
			}
		})
	}
}
