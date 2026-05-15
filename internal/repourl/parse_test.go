package repourl_test

import (
	"errors"
	"testing"

	"git-clone-manager/internal/repourl"
)

func TestParseHTTPSURL(t *testing.T) {
	parts, err := repourl.Parse("https://github.com/nWave-ai/nWave")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parts.Hostname != "github.com" {
		t.Fatalf("Hostname = %q, want %q", parts.Hostname, "github.com")
	}

	if parts.PathPrefix != "nWave-ai" {
		t.Fatalf("PathPrefix = %q, want %q", parts.PathPrefix, "nWave-ai")
	}

	if parts.RepositoryName != "nWave" {
		t.Fatalf("RepositoryName = %q, want %q", parts.RepositoryName, "nWave")
	}
}

func TestParseGitAtURL(t *testing.T) {
	parts, err := repourl.Parse("git@github.com:nWave-ai/nWave.git")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if parts.Hostname != "github.com" {
		t.Fatalf("Hostname = %q, want %q", parts.Hostname, "github.com")
	}

	if parts.PathPrefix != "nWave-ai" {
		t.Fatalf("PathPrefix = %q, want %q", parts.PathPrefix, "nWave-ai")
	}

	if parts.RepositoryName != "nWave" {
		t.Fatalf("RepositoryName = %q, want %q", parts.RepositoryName, "nWave")
	}
}

func TestParseSupportedURLFormats(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		hostname string
		prefix   string
		repoName string
	}{
		{
			name:     "ssh url",
			rawURL:   "ssh://git@github.com/nWave-ai/nWave",
			hostname: "github.com",
			prefix:   "nWave-ai",
			repoName: "nWave",
		},
		{
			name:     "deep path prefix",
			rawURL:   "https://gitlab.com/group/subgroup/repo",
			hostname: "gitlab.com",
			prefix:   "group/subgroup",
			repoName: "repo",
		},
		{
			name:     "no path prefix",
			rawURL:   "https://example.com/repo",
			hostname: "example.com",
			prefix:   "",
			repoName: "repo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parts, err := repourl.Parse(test.rawURL)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}

			if parts.Hostname != test.hostname {
				t.Fatalf("Hostname = %q, want %q", parts.Hostname, test.hostname)
			}

			if parts.PathPrefix != test.prefix {
				t.Fatalf("PathPrefix = %q, want %q", parts.PathPrefix, test.prefix)
			}

			if parts.RepositoryName != test.repoName {
				t.Fatalf("RepositoryName = %q, want %q", parts.RepositoryName, test.repoName)
			}
		})
	}
}

func TestParseMalformedURLReturnsParseError(t *testing.T) {
	_, err := repourl.Parse("https://github.com")
	if err == nil {
		t.Fatal("Parse unexpectedly succeeded")
	}

	var parseErr *repourl.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("Parse error type = %T, want *repourl.ParseError", err)
	}
}
