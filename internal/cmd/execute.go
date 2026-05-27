package cmd

import (
	"errors"
	"fmt"
	"io"

	"git-clone-manager/internal/exitcodes"

	"github.com/spf13/cobra"
)

var errUnknownCommand = errors.New("unknown command")

func Execute(args []string, stdout, stderr io.Writer) int {
	return execute(args, stdout, stderr, DefaultDependencies())
}

func execute(args []string, stdout, stderr io.Writer, deps Dependencies) int {
	root := newRootCommand(deps)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	err := detectUnknownCommand(root, args)
	if err == nil {
		err = root.Execute()
	}
	err = normalizeError(err)
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

	if errors.Is(err, errUnknownCommand) {
		return exitcodes.UsageError(err)
	}

	return err
}

func detectUnknownCommand(root *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	_, _, err := root.Find(args)
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w: %v", errUnknownCommand, err)
}
