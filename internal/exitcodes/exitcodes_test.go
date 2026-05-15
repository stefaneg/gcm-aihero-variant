package exitcodes_test

import (
	"errors"
	"testing"

	"git-clone-manager/internal/exitcodes"
)

func TestCodeMapsSuccessGeneralAndUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "success", err: nil, want: exitcodes.Success},
		{name: "general", err: errors.New("boom"), want: exitcodes.General},
		{name: "usage", err: exitcodes.UsageError(errors.New("bad args")), want: exitcodes.Usage},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := exitcodes.Code(test.err); got != test.want {
				t.Fatalf("Code(%v) = %d, want %d", test.err, got, test.want)
			}
		})
	}
}
