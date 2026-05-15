package exitcodes

import "errors"

const (
	Success = 0
	General = 1
	Usage   = 2
)

type Error struct {
	code int
	err  error
}

func (errorWithCode *Error) Error() string {
	return errorWithCode.err.Error()
}

func (errorWithCode *Error) Unwrap() error {
	return errorWithCode.err
}

func (errorWithCode *Error) ExitCode() int {
	return errorWithCode.code
}

func WithCode(code int, err error) error {
	if err == nil {
		return nil
	}

	return &Error{code: code, err: err}
}

func UsageError(err error) error {
	return WithCode(Usage, err)
}

func Code(err error) int {
	if err == nil {
		return Success
	}

	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}

	return General
}
