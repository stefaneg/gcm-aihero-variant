package derivedpath_test

import (
	"testing"

	"git-clone-manager/internal/derivedpath"
)

func TestDerive(t *testing.T) {
	tests := []struct {
		name        string
		cloneRoot   string
		hostname    string
		pathPrefix  string
		repository  string
		wantDerived string
	}{
		{
			name:        "with path prefix",
			cloneRoot:   "~/src",
			hostname:    "github.com",
			pathPrefix:  "nWave-ai",
			repository:  "nWave",
			wantDerived: "~/src/github.com/nWave-ai/nWave",
		},
		{
			name:        "without path prefix",
			cloneRoot:   "~/src",
			hostname:    "example.com",
			pathPrefix:  "",
			repository:  "repo",
			wantDerived: "~/src/example.com/repo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := derivedpath.Derive(test.cloneRoot, test.hostname, test.pathPrefix, test.repository)
			if got != test.wantDerived {
				t.Fatalf("Derive() = %q, want %q", got, test.wantDerived)
			}
		})
	}
}
