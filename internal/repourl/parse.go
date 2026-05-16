package repourl

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

type Parts struct {
	Hostname       string
	PathPrefix     string
	RepositoryName string
}

type ParseError struct {
	RawURL string
	Reason string
}

func (errorWithContext *ParseError) Error() string {
	return fmt.Sprintf("parse repository URL %q: %s", errorWithContext.RawURL, errorWithContext.Reason)
}

func Parse(rawURL string) (Parts, error) {
	if strings.Contains(rawURL, "@") && strings.Contains(rawURL, ":") && !strings.Contains(rawURL, "://") {
		hostAndPath := strings.SplitN(rawURL, ":", 2)
		if len(hostAndPath) != 2 {
			return Parts{}, &ParseError{RawURL: rawURL, Reason: "missing repository path"}
		}

		host := hostAndPath[0]
		path := hostAndPath[1]

		atIndex := strings.LastIndex(host, "@")
		if atIndex == -1 || atIndex == len(host)-1 {
			return Parts{}, &ParseError{RawURL: rawURL, Reason: "missing hostname"}
		}

		return parseParts(rawURL, host[atIndex+1:], path)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return Parts{}, &ParseError{RawURL: rawURL, Reason: err.Error()}
	}

	if parsedURL.Hostname() == "" {
		return Parts{}, &ParseError{RawURL: rawURL, Reason: "missing hostname"}
	}

	return parseParts(rawURL, parsedURL.Hostname(), parsedURL.Path)
}

func parseParts(rawURL, hostname, rawPath string) (Parts, error) {
	path := strings.Trim(strings.TrimSuffix(rawPath, ".git"), "/")
	if path == "" {
		return Parts{}, &ParseError{RawURL: rawURL, Reason: "missing repository path"}
	}

	segments := strings.Split(path, "/")
	repositoryName := segments[len(segments)-1]
	if repositoryName == "" {
		return Parts{}, &ParseError{RawURL: rawURL, Reason: "missing repository name"}
	}

	parts := Parts{
		Hostname:       hostname,
		RepositoryName: repositoryName,
	}

	if len(segments) > 1 {
		parts.PathPrefix = strings.Join(segments[:len(segments)-1], "/")
	}

	return parts, nil
}

func (p Parts) DerivedPath(cloneRoot string) string {
	if p.PathPrefix == "" {
		return filepath.Join(cloneRoot, p.Hostname, p.RepositoryName)
	}

	return filepath.Join(cloneRoot, p.Hostname, p.PathPrefix, p.RepositoryName)
}
