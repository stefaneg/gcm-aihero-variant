# 0007 — Status collector

## What to build

A Status Collector module that gathers the status of a single repository using the Git Runner interface. This module is called once per repository by the Worker Pool, so it must be safe for concurrent use.

For each repository it collects:
- Current branch name
- Default branch name (from `refs/remotes/origin/HEAD`; falls back to `main` when not set)
- Number of commits the local default branch is behind the remote
- Count of dirty files (uncommitted local changes, tracked and untracked)
- Error state: `fetch-failed` when the fetch fails, `no-remote` when no remote is configured

When `--no-fetch` is active, the fetch step is skipped entirely and commits-behind is computed from local state only.

The Status Collector returns a typed result struct per repository. It does not sort, format, or filter results.

## Acceptance criteria

- [ ] Returns correct branch, default branch, commits-behind, and dirty count for a clean on-default-branch repository
- [ ] Returns correct non-default branch name when checked out on a feature branch
- [ ] Falls back to `main` as default branch when `refs/remotes/origin/HEAD` is not set
- [ ] Returns `fetch-failed` error state (not a fatal error) when git fetch fails
- [ ] Returns `no-remote` error state when the repository has no remote configured
- [ ] With `--no-fetch`: no fetch is performed; commits-behind reflects local state
- [ ] All tests use the fake Git Runner from 0004 — no real git processes or filesystem I/O in unit tests
- [ ] Concurrent calls with the same fake runner do not race (verified with `-race` flag)

## Blocked by

- 0004 — git runner
- 0006 — repository walker and worker pool
