package statusformatter

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"git-clone-manager/internal/statuscollector"
)

type Options struct {
	StdoutIsTTY          bool
	NoColor              bool
	NonDefaultOnly       bool
	TotalRepositoryCount int
}

type formattedRow struct {
	relativePath string
	result       statuscollector.Result
}

func Format(cloneRoot string, results []statuscollector.Result, options Options) (string, error) {
	rows := make([]formattedRow, 0, len(results))
	for _, result := range results {
		relativePath, err := filepath.Rel(cloneRoot, result.RepositoryPath)
		if err != nil {
			return "", fmt.Errorf("compute repository path relative to clone root: %w", err)
		}

		rows = append(rows, formattedRow{
			relativePath: relativePath,
			result:       result,
		})
	}

	sortRows(rows)

	pathWidth := len("path")
	branchWidth := len("branch")
	behindWidth := 1
	dirtyWidth := 1
	for _, row := range rows {
		if len(row.relativePath) > pathWidth {
			pathWidth = len(row.relativePath)
		}
		if len(row.result.CurrentBranch) > branchWidth {
			branchWidth = len(row.result.CurrentBranch)
		}
		if digitWidth(row.result.CommitsBehind) > behindWidth {
			behindWidth = digitWidth(row.result.CommitsBehind)
		}
		if digitWidth(row.result.DirtyCount) > dirtyWidth {
			dirtyWidth = digitWidth(row.result.DirtyCount)
		}
	}

	var builder strings.Builder
	builder.WriteString("Repos under ")
	builder.WriteString(cloneRoot)
	builder.WriteString(":\n")

	totalRepositoryCount := options.TotalRepositoryCount
	if totalRepositoryCount == 0 {
		totalRepositoryCount = len(rows)
	}

	if options.NonDefaultOnly && len(rows) == 0 && totalRepositoryCount > 0 {
		builder.WriteString("No repositories on non-default branches.\n")
	} else {
		for _, row := range rows {
			badges := formatBadges(row.result, options)
			builder.WriteString(fmt.Sprintf(
				"%-*s  %-*s  behind=%*d  dirty=%*d",
				pathWidth,
				row.relativePath,
				branchWidth,
				row.result.CurrentBranch,
				behindWidth,
				row.result.CommitsBehind,
				dirtyWidth,
				row.result.DirtyCount,
			))
			if badges != "" {
				builder.WriteString("  ")
				builder.WriteString(badges)
			}
			builder.WriteByte('\n')
		}
	}

	currentCount, behindCount, nonDefaultCount := summarize(rows)
	if options.NonDefaultOnly && len(rows) == 0 && totalRepositoryCount > 0 {
		builder.WriteString(fmt.Sprintf("%d repositories, 0 non-default.\n", totalRepositoryCount))
	} else {
		builder.WriteString(fmt.Sprintf(
			"%d repos — %d current, %d behind, %d non-default-branch\n",
			len(rows),
			currentCount,
			behindCount,
			nonDefaultCount,
		))
		if len(rows) > 0 && !options.NonDefaultOnly {
			builder.WriteString("Tips: gcm status --non-default\n")
		}
	}

	return builder.String(), nil
}

func digitWidth(value int) int {
	return len(strconv.Itoa(value))
}

func sortRows(rows []formattedRow) {
	slices.SortFunc(rows, func(left formattedRow, right formattedRow) int {
		leftNonDefault := isNonDefault(left.result)
		rightNonDefault := isNonDefault(right.result)
		if leftNonDefault != rightNonDefault {
			if leftNonDefault {
				return -1
			}
			return 1
		}

		leftTier := sortTier(left.result)
		rightTier := sortTier(right.result)
		if leftTier != rightTier {
			return leftTier - rightTier
		}

		if left.result.CommitsBehind != right.result.CommitsBehind {
			return right.result.CommitsBehind - left.result.CommitsBehind
		}

		return strings.Compare(left.relativePath, right.relativePath)
	})
}

func summarize(rows []formattedRow) (currentCount int, behindCount int, nonDefaultCount int) {
	for _, row := range rows {
		switch {
		case isNonDefault(row.result):
			nonDefaultCount++
		case row.result.ErrorState != statuscollector.ErrorStateNone:
			continue
		case row.result.CommitsBehind > 0:
			behindCount++
		default:
			currentCount++
		}
	}

	return currentCount, behindCount, nonDefaultCount
}

func formatBadges(result statuscollector.Result, options Options) string {
	var badges []string

	if result.CommitsBehind > 0 {
		badges = append(badges, maybeColorize("[behind]", ansiYellow, shouldColorize(options)))
	}
	if isNonDefault(result) {
		badges = append(badges, maybeColorize(fmt.Sprintf("[!%s]", result.DefaultBranch), ansiBlue, shouldColorize(options)))
	}
	if result.ErrorState == statuscollector.ErrorStateFetchFailed {
		badges = append(badges, maybeColorize("[fetch-failed]", ansiRed, shouldColorize(options)))
	}
	if result.ErrorState == statuscollector.ErrorStateNoRemote {
		badges = append(badges, maybeColorize("[no-remote]", ansiMagenta, shouldColorize(options)))
	}
	if result.ErrorState == statuscollector.ErrorStateDefaultUnknown {
		badges = append(badges, maybeColorize("[default-unknown]", ansiMagenta, shouldColorize(options)))
	}
	if result.ErrorState == statuscollector.ErrorStateUnknown {
		badges = append(badges, maybeColorize("[error]", ansiRed, shouldColorize(options)))
	}

	return strings.Join(badges, " ")
}

func isNonDefault(result statuscollector.Result) bool {
	if result.ErrorState == statuscollector.ErrorStateDefaultUnknown || result.ErrorState == statuscollector.ErrorStateNoRemote {
		return false
	}
	if result.DefaultBranch == "" {
		return false
	}
	return result.CurrentBranch != result.DefaultBranch
}

func sortTier(result statuscollector.Result) int {
	if isNonDefault(result) {
		return 0
	}
	if result.ErrorState == statuscollector.ErrorStateNoRemote || result.ErrorState == statuscollector.ErrorStateDefaultUnknown || result.ErrorState == statuscollector.ErrorStateUnknown {
		return 2
	}
	if result.ErrorState == statuscollector.ErrorStateFetchFailed && result.DefaultBranch == "" {
		return 2
	}
	return 1
}

func maybeColorize(text string, colorCode string, colorize bool) string {
	if !colorize {
		return text
	}

	return colorCode + text + ansiReset
}

func shouldColorize(options Options) bool {
	return options.StdoutIsTTY && !options.NoColor
}

const (
	ansiReset   = "\x1b[0m"
	ansiRed     = "\x1b[31m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
)
