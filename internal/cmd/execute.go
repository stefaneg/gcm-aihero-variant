package cmd

import (
	"fmt"
	"io"
	"strings"

	"git-clone-manager/internal/exitcodes"
)

func Execute(args []string, stdout, stderr io.Writer) int {
	root := NewRootCommand()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	err := normalizeError(root.Execute())
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
	}

	return exitcodes.Code(err)
}

func normalizeError(err error) error {
	if err == nil {
		return nil
	}

	if exitcodes.Code(err) != exitcodes.General {
		return err
	}

	message := err.Error()
	if strings.HasPrefix(message, "unknown command ") {
		return exitcodes.UsageError(err)
	}

	return err
}
