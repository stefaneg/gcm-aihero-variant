package statusformatter_test

import (
	"strings"
	"testing"

	"git-clone-manager/internal/statuscollector"
	"git-clone-manager/internal/statusformatter"
)

func TestFormatBuildsPlainTextStatusTable(t *testing.T) {
	cloneRoot := "/repos"
	results := []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/beta",
			CurrentBranch:  "feature/beta",
			DefaultBranch:  "main",
			CommitsBehind:  5,
			DirtyCount:     0,
		},
		{
			RepositoryPath: "/repos/github.com/acme/current",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  0,
			DirtyCount:     0,
		},
		{
			RepositoryPath: "/repos/github.com/acme/feature",
			CurrentBranch:  "feature/login",
			DefaultBranch:  "main",
			CommitsBehind:  2,
			DirtyCount:     1,
		},
		{
			RepositoryPath: "/repos/github.com/acme/behind",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  3,
			DirtyCount:     0,
		},
		{
			RepositoryPath: "/repos/github.com/acme/alpha",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  3,
			DirtyCount:     2,
		},
	}

	got, err := statusformatter.Format(cloneRoot, results, statusformatter.Options{})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	want := "" +
		"Repos under /repos:\n" +
		"github.com/acme/beta     feature/beta   behind=5  dirty=0  [behind] [!main]\n" +
		"github.com/acme/feature  feature/login  behind=2  dirty=1  [behind] [!main]\n" +
		"github.com/acme/alpha    main           behind=3  dirty=2  [behind]\n" +
		"github.com/acme/behind   main           behind=3  dirty=0  [behind]\n" +
		"github.com/acme/current  main           behind=0  dirty=0\n" +
		"5 repos — 1 current, 2 behind, 2 non-default-branch\n" +
		"Tips: gcm status --non-default\n"

	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormatShowsErrorBadgesAndExcludesThemFromBehindAndCurrentCounts(t *testing.T) {
	cloneRoot := "/repos"
	results := []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/non-default",
			CurrentBranch:  "feature/search",
			DefaultBranch:  "master",
			CommitsBehind:  4,
			DirtyCount:     0,
		},
		{
			RepositoryPath: "/repos/github.com/acme/fetch-failed",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  6,
			DirtyCount:     1,
			ErrorState:     statuscollector.ErrorStateFetchFailed,
		},
		{
			RepositoryPath: "/repos/github.com/acme/no-remote",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  0,
			DirtyCount:     2,
			ErrorState:     statuscollector.ErrorStateNoRemote,
		},
		{
			RepositoryPath: "/repos/github.com/acme/default-unknown",
			CurrentBranch:  "main",
			CommitsBehind:  0,
			DirtyCount:     0,
			ErrorState:     statuscollector.ErrorStateDefaultUnknown,
		},
		{
			RepositoryPath: "/repos/github.com/acme/current",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  0,
			DirtyCount:     0,
		},
	}

	got, err := statusformatter.Format(cloneRoot, results, statusformatter.Options{})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if !strings.Contains(got, "[!master]") {
		t.Fatalf("Format() = %q, want [!master] badge", got)
	}

	if strings.Contains(got, "[!main]") {
		t.Fatalf("Format() = %q, did not want [!main] badge", got)
	}

	if !strings.Contains(got, "[fetch-failed]") {
		t.Fatalf("Format() = %q, want [fetch-failed] badge", got)
	}

	if !strings.Contains(got, "[no-remote]") {
		t.Fatalf("Format() = %q, want [no-remote] badge", got)
	}

	if !strings.Contains(got, "[default-unknown]") {
		t.Fatalf("Format() = %q, want [default-unknown] badge", got)
	}

	if !strings.Contains(got, "5 repos — 1 current, 0 behind, 1 non-default-branch") {
		t.Fatalf("Format() = %q, want error states excluded from current/behind counts", got)
	}
}

func TestFormatSortsDefaultUnknownInIncompleteDataTier(t *testing.T) {
	cloneRoot := "/repos"
	results := []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/unknown",
			CurrentBranch:  "feature/unknown",
			CommitsBehind:  0,
			DirtyCount:     0,
			ErrorState:     statuscollector.ErrorStateDefaultUnknown,
		},
		{
			RepositoryPath: "/repos/github.com/acme/current",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  0,
			DirtyCount:     0,
		},
		{
			RepositoryPath: "/repos/github.com/acme/no-remote",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  0,
			DirtyCount:     0,
			ErrorState:     statuscollector.ErrorStateNoRemote,
		},
		{
			RepositoryPath: "/repos/github.com/acme/feature",
			CurrentBranch:  "feature/login",
			DefaultBranch:  "main",
			CommitsBehind:  0,
			DirtyCount:     0,
		},
	}

	got, err := statusformatter.Format(cloneRoot, results, statusformatter.Options{})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	feature := "github.com/acme/feature"
	current := "github.com/acme/current"
	noRemote := "github.com/acme/no-remote"
	unknown := "github.com/acme/unknown"

	if strings.Index(got, feature) > strings.Index(got, current) {
		t.Fatalf("Format() = %q, want non-default row above healthy default row", got)
	}
	if strings.Index(got, current) > strings.Index(got, noRemote) {
		t.Fatalf("Format() = %q, want healthy default row above incomplete rows", got)
	}
	if strings.Index(got, noRemote) > strings.Index(got, unknown) {
		t.Fatalf("Format() = %q, want incomplete rows sorted alphabetically", got)
	}
	if strings.Contains(got, "[!]") {
		t.Fatalf("Format() = %q, did not want non-default badge for unknown default branch", got)
	}
}

func TestFormatShowsFilterAwareEmptyNonDefaultMessage(t *testing.T) {
	got, err := statusformatter.Format("/repos", nil, statusformatter.Options{
		NonDefaultOnly:       true,
		TotalRepositoryCount: 42,
	})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	want := "" +
		"Repos under /repos:\n" +
		"No repositories on non-default branches.\n" +
		"42 repositories, 0 non-default.\n"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormatLeavesFullyEmptyCloneRootUnchangedForNonDefaultFilter(t *testing.T) {
	got, err := statusformatter.Format("/repos", nil, statusformatter.Options{
		NonDefaultOnly: true,
	})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	want := "" +
		"Repos under /repos:\n" +
		"0 repos — 0 current, 0 behind, 0 non-default-branch\n"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormatDropsPullTipButKeepsNonDefaultTipForPopulatedStatus(t *testing.T) {
	cloneRoot := "/repos"
	results := []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/current",
			CurrentBranch:  "main",
			DefaultBranch:  "main",
			CommitsBehind:  0,
			DirtyCount:     0,
		},
	}

	got, err := statusformatter.Format(cloneRoot, results, statusformatter.Options{})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if strings.Contains(got, "gcm pull") {
		t.Fatalf("Format() = %q, did not want gcm pull tip", got)
	}
	if !strings.Contains(got, "Tips: gcm status --non-default") {
		t.Fatalf("Format() = %q, want non-default tip", got)
	}
}

func TestFormatAppliesColorOnlyForTTYWithoutNoColor(t *testing.T) {
	cloneRoot := "/repos"
	results := []statuscollector.Result{
		{
			RepositoryPath: "/repos/github.com/acme/feature",
			CurrentBranch:  "feature/login",
			DefaultBranch:  "main",
			CommitsBehind:  2,
			DirtyCount:     1,
		},
	}

	colorized, err := statusformatter.Format(cloneRoot, results, statusformatter.Options{
		StdoutIsTTY: true,
	})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if !strings.Contains(colorized, "\x1b[") {
		t.Fatalf("Format() = %q, want ANSI color escapes when stdout is a TTY", colorized)
	}

	plain, err := statusformatter.Format(cloneRoot, results, statusformatter.Options{
		StdoutIsTTY: false,
	})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if strings.Contains(plain, "\x1b[") {
		t.Fatalf("Format() = %q, did not want ANSI color escapes for non-TTY output", plain)
	}

	noColor, err := statusformatter.Format(cloneRoot, results, statusformatter.Options{
		StdoutIsTTY: true,
		NoColor:     true,
	})
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if strings.Contains(noColor, "\x1b[") {
		t.Fatalf("Format() = %q, did not want ANSI color escapes when NO_COLOR is set", noColor)
	}
}
