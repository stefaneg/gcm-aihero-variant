# 0013 — Default branch unknown handling, `[default-unknown]` badge, three-tier sort

Status: done

## What to build

Replace the PRD's original "fall back to `main`" behaviour for default-branch detection with an honest "unknown" treatment, surface it as a distinct badge, and extend the status table sort to a three-tier hierarchy.

**Status Collector**. When `refs/remotes/origin/HEAD` is unset on a repository that has an `origin` remote, the default branch is reported as **unknown** — no `main` fallback. The result struct gains a `default-unknown` error state alongside the existing `fetch-failed` and `no-remote` states. The non-default-branch comparison cannot be performed when the default is unknown, so it is skipped for that row.

**Status Formatter**.

Recognised badges now include `[default-unknown]`. A row with `[default-unknown]` does **not** receive a `[!{default_branch}]` badge — the two are mutually exclusive by definition.

Sort order becomes three tiers, in order top to bottom:

1. Non-default-branch rows (active work)
2. Healthy default-branch rows (including `[behind]`)
3. Incomplete-data rows: `[no-remote]` and `[default-unknown]`

Within each tier, sort by commits-behind descending, then alphabetical by derived path. `[fetch-failed]` rows continue to be placed in the tier that matches what is known about them (e.g., behind/non-default if those are known; otherwise the incomplete-data tier).

Summary counts: `[no-remote]` and `[default-unknown]` rows continue to be excluded from `current` and `behind` counts.

## Acceptance criteria

- [ ] A repo with no `refs/remotes/origin/HEAD` but a working `origin` remote is reported as `default-unknown`, not as being on a non-default branch named `main`
- [ ] `[default-unknown]` rows render with the badge in the badges column and no `[!{default_branch}]` badge
- [ ] Three-tier sort: a `[default-unknown]` row sorts below a healthy `main` row; a non-default-branch row sorts above both
- [ ] Within the incomplete-data tier, `[no-remote]` and `[default-unknown]` rows interleave alphabetically by derived path
- [ ] `current` and `behind` summary counts exclude `[default-unknown]` rows
- [ ] Status Collector unit tests cover the HEAD-unset case and assert the unknown state
- [ ] Status Formatter unit tests cover the three-tier sort and the new badge
- [ ] Integration test exercises a repo with HEAD intentionally unset and asserts the end-to-end status output

## Blocked by

- None
