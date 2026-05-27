package cmd

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"git-clone-manager/internal/configstore"
	"git-clone-manager/internal/exitcodes"
	"git-clone-manager/internal/statuscollector"
)

type fakeStatusCollector struct {
	results []statuscollector.Result
	err     error

	cloneRoot string
	noFetch   bool
}

func (collector *fakeStatusCollector) Collect(cloneRoot string, noFetch bool) ([]statuscollector.Result, error) {
	collector.cloneRoot = cloneRoot
	collector.noFetch = noFetch
	return collector.results, collector.err
}

func executeStatusCommand(cloneRoot string, collector *fakeStatusCollector, args []string, stdout, stderr *bytes.Buffer) int {
	deps := DefaultDependencies()
	deps.LoadEffectiveStatusConfig = func() (configstore.EffectiveConfig, error) {
		return configstore.EffectiveConfig{CloneRoot: cloneRoot}, nil
	}
	deps.NewStatusCollector = func() statusCollector {
		return collector
	}

	return execute(args, stdout, stderr, deps)
}

func TestStatusShowsFormattedTableForRepositoriesUnderCloneRoot(t *testing.T) {
	t.Parallel()

	cloneRoot := "/repos"
	collector := &fakeStatusCollector{results: []statuscollector.Result{
		{RepositoryPath: "/repos/github.com/acme/current", CurrentBranch: "main", DefaultBranch: "main"},
		{RepositoryPath: "/repos/github.com/acme/behind", CurrentBranch: "main", DefaultBranch: "main", CommitsBehind: 1},
		{RepositoryPath: "/repos/github.com/acme/feature", CurrentBranch: "feature/login", DefaultBranch: "main", CommitsBehind: 1, DirtyCount: 1},
	}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	status := stdout.String()
	for _, want := range []string{
		"Repos under /repos:\n",
		"github.com/acme/feature  feature/login  behind=1  dirty=1  [behind] [!main]",
		"github.com/acme/behind   main           behind=1  dirty=0  [behind]",
		"github.com/acme/current  main           behind=0  dirty=0",
		"3 repos - 1 current, 1 behind, 1 non-default-branch",
		"Tips: gcm status --non-default",
	} {
		if !strings.Contains(strings.ReplaceAll(status, "—", "-"), want) {
			t.Fatalf("status output = %q, want %q", status, want)
		}
	}
}

func TestStatusNonDefaultFiltersTable(t *testing.T) {
	cloneRoot := "/repos"
	collector := &fakeStatusCollector{results: []statuscollector.Result{
		{RepositoryPath: "/repos/github.com/acme/current", CurrentBranch: "main", DefaultBranch: "main"},
		{RepositoryPath: "/repos/github.com/acme/feature", CurrentBranch: "feature/login", DefaultBranch: "main"},
	}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status", "--non-default"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	status := stdout.String()
	if !strings.Contains(status, "github.com/acme/feature  feature/login  behind=0  dirty=0  [!main]") {
		t.Fatalf("status output = %q, want non-default row", status)
	}
	if strings.Contains(status, "github.com/acme/current") {
		t.Fatalf("status output = %q, did not want default-branch row", status)
	}
}

func TestStatusNonDefaultFormatsWidthsFromFilteredRows(t *testing.T) {
	cloneRoot := "/repos"
	collector := &fakeStatusCollector{results: []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/default-stale",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  408,
			DirtyCount:     77,
		},
		{
			RepositoryPath: "/repos/github.com/acme/feature",
			CurrentBranch:  "feature/login",
			DefaultBranch:  "main",
			CommitsBehind:  9,
			DirtyCount:     1,
		},
	}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status", "--non-default"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	status := stdout.String()
	if !strings.Contains(status, "github.com/acme/feature  feature/login  behind=9  dirty=1  [behind] [!main]") {
		t.Fatalf("status output = %q, want filtered row without excluded-row numeric padding", status)
	}
	if strings.Contains(status, "behind=  9") || strings.Contains(status, "dirty= 1") {
		t.Fatalf("status output = %q, did not want numeric widths from filtered-out row", status)
	}
}

func TestStatusNonDefaultShowsFilterAwareEmptyMessage(t *testing.T) {
	cloneRoot := "/repos"
	collector := &fakeStatusCollector{results: []statuscollector.Result{
		{RepositoryPath: "/repos/github.com/acme/current", CurrentBranch: "main", DefaultBranch: "main"},
	}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status", "--non-default"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	status := stdout.String()
	for _, want := range []string{"No repositories on non-default branches.", "1 repositories, 0 non-default."} {
		if !strings.Contains(status, want) {
			t.Fatalf("status output = %q, want %q", status, want)
		}
	}
	if strings.Contains(status, "Tips:") {
		t.Fatalf("status output = %q, did not want tip", status)
	}
}

func TestStatusNoFetchUsesLocalStateWhenRemoteIsUnreachable(t *testing.T) {
	cloneRoot := "/repos"
	collector := &fakeStatusCollector{results: []statuscollector.Result{
		{RepositoryPath: "/repos/github.com/acme/current", CurrentBranch: "main", DefaultBranch: "main"},
	}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status", "--no-fetch"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if !collector.noFetch {
		t.Fatalf("collector noFetch = false, want true")
	}
	if strings.Contains(stdout.String(), "[fetch-failed]") {
		t.Fatalf("status output = %q, did not want fetch-failed marker", stdout.String())
	}
}

func TestStatusReturnsNonZeroAndShowsFetchFailedRepositories(t *testing.T) {
	cloneRoot := "/repos"
	collector := &fakeStatusCollector{results: []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/offline",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			ErrorState:     statuscollector.ErrorStateFetchFailed,
		},
	}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status"}, &stdout, &stderr)
	if exitCode != exitcodes.General {
		t.Fatalf("Execute exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", exitCode, exitcodes.General, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), "[fetch-failed]") {
		t.Fatalf("stdout = %q, want fetch-failed marker", stdout.String())
	}
	if !strings.Contains(stderr.String(), "one or more repositories failed to fetch") {
		t.Fatalf("stderr = %q, want partial-failure message", stderr.String())
	}
}

func TestStatusShowsIncompleteDataRepositoriesWithoutFailing(t *testing.T) {
	cloneRoot := "/repos"
	collector := &fakeStatusCollector{results: []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/local-only",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			ErrorState:     statuscollector.ErrorStateNoRemote,
		},
		{
			RepositoryPath: "/repos/github.com/acme/unknown-default",
			CurrentBranch:  "main",
			ErrorState:     statuscollector.ErrorStateDefaultUnknown,
		},
	}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	status := stdout.String()
	for _, want := range []string{"[no-remote]", "[default-unknown]", "2 repos"} {
		if !strings.Contains(status, want) {
			t.Fatalf("status output = %q, want %q", status, want)
		}
	}
	if strings.Contains(status, "[!main]") {
		t.Fatalf("status output = %q, did not want fallback main non-default marker", status)
	}
}

func TestStatusPropagatesHardCollectionErrors(t *testing.T) {
	cloneRoot := filepath.Join(t.TempDir(), "src")
	collector := &fakeStatusCollector{err: errors.New("walk failed")}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status"}, &stdout, &stderr)
	if exitCode != exitcodes.General {
		t.Fatalf("Execute exit code = %d, want %d", exitCode, exitcodes.General)
	}
	if !strings.Contains(stderr.String(), "walk failed") {
		t.Fatalf("stderr = %q, want hard error", stderr.String())
	}
}

func TestStatusRendersPartialResultsBeforeReturningBatchError(t *testing.T) {
	cloneRoot := "/repos"
	collector := &fakeStatusCollector{
		results: []statuscollector.Result{
			{RepositoryPath: "/repos/github.com/acme/current", CurrentBranch: "main", DefaultBranch: "main"},
			{RepositoryPath: "/repos/github.com/acme/broken", ErrorState: statuscollector.ErrorStateUnknown},
		},
		err: errors.New("one or more repositories could not be inspected"),
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status"}, &stdout, &stderr)
	if exitCode != exitcodes.General {
		t.Fatalf("Execute exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", exitCode, exitcodes.General, stdout.String(), stderr.String())
	}

	if !strings.Contains(stdout.String(), "github.com/acme/current") || !strings.Contains(stdout.String(), "[error]") {
		t.Fatalf("stdout = %q, want partial table with error badge", stdout.String())
	}
	if !strings.Contains(stderr.String(), "one or more repositories could not be inspected") {
		t.Fatalf("stderr = %q, want batch error", stderr.String())
	}
}

func TestStatusCommandUsesExpandedCloneRoot(t *testing.T) {
	collector := &fakeStatusCollector{}
	t.Setenv("HOME", "/home/dev")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand("~/src", collector, []string{"status"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	if collector.cloneRoot != filepath.Join("/home/dev", "src") {
		t.Fatalf("collector cloneRoot = %q, want expanded home path", collector.cloneRoot)
	}
}

func TestStatusCommandKeepsFetchFailedWhenNonDefaultFilterExcludesRow(t *testing.T) {
	cloneRoot := "/repos"
	collector := &fakeStatusCollector{results: []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/offline",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			ErrorState:     statuscollector.ErrorStateFetchFailed,
		},
	}}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := executeStatusCommand(cloneRoot, collector, []string{"status", "--non-default"}, &stdout, &stderr)
	if exitCode != exitcodes.General {
		t.Fatalf("Execute exit code = %d, want %d", exitCode, exitcodes.General)
	}
	if !strings.Contains(stderr.String(), "one or more repositories failed to fetch") {
		t.Fatalf("stderr = %q, want partial-failure message", stderr.String())
	}
}

func TestStatusResultIsNonDefaultSkipsIncompleteData(t *testing.T) {
	for _, result := range []statuscollector.Result{
		{CurrentBranch: "feature", ErrorState: statuscollector.ErrorStateNoRemote},
		{CurrentBranch: "feature", ErrorState: statuscollector.ErrorStateDefaultUnknown},
		{CurrentBranch: "feature"},
	} {
		if statusResultIsNonDefault(result) {
			t.Fatalf("statusResultIsNonDefault(%#v) = true, want false", result)
		}
	}

	if !statusResultIsNonDefault(statuscollector.Result{CurrentBranch: "feature", DefaultBranch: "main"}) {
		t.Fatal("statusResultIsNonDefault(feature/main) = false, want true")
	}
}
