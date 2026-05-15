# 0008 — Status formatter

## What to build

A Status Formatter module that takes a slice of Status Collector results and produces the formatted status table string. No git operations, no filesystem access — pure transformation of data into text.

**Sort order**: non-default-branch repositories first, then by commits-behind descending, then alphabetical by derived path.

**Table structure**:
- Header line: `Repos under <clone_root>:`
- One row per repository: relative path, current branch, commits behind, dirty file count, badges
- Summary line: `N repos — M current, P behind, Q non-default-branch`
- Tips: exact command strings for `gcm pull` and `gcm status --non-default`

**Badges** (text, always present):
- `[behind]` — repository is behind its remote
- `[!{default_branch}]` — repository is on a non-default branch, using that repository's actual default branch name (e.g., `[!master]`, `[!develop]`)
- `[fetch-failed]` — fetch failed; excluded from behind/current counts
- `[no-remote]` — no remote configured; excluded from behind/current counts

**Colour**: applied on top of text badges when stdout is a TTY and `NO_COLOR` is not set. Badges are always present as text regardless of colour support — colour is a progressive enhancement only.

## Acceptance criteria

- [ ] Non-default-branch repositories sort before behind on-default-branch repositories
- [ ] Within the non-default-branch group: sorts by commits-behind descending, then alphabetical
- [ ] Within the behind group: sorts by commits-behind descending, then alphabetical
- [ ] Clean on-default-branch repositories sort last, alphabetically
- [ ] `[behind]` badge appears on repositories with commits-behind > 0
- [ ] `[!master]` badge appears (not `[!main]`) when the default branch is `master`
- [ ] `[fetch-failed]` and `[no-remote]` repositories appear in the table but are excluded from summary counts
- [ ] Summary counts are correct: current excludes behind, non-default-branch, fetch-failed, and no-remote
- [ ] Tips line references exact command strings: `gcm pull` and `gcm status --non-default`
- [ ] With a non-TTY stdout (simulated in test), no colour escape codes appear in output
- [ ] With `NO_COLOR` set, no colour escape codes appear even on a TTY
- [ ] Unit tests cover all sort order cases and all badge combinations

## Blocked by

- 0007 — status collector
