# 0009 — `gcm status` end-to-end

## What to build

The complete `gcm status` command wired through all layers: Config Store (clone root) → Repository Walker → Worker Pool → Status Collector → Status Formatter → stdout.

This slice wires the modules together, adds the `--no-fetch` and `--non-default` flags, enforces the exit code contract, and verifies end-to-end behaviour with integration tests against real local git repositories.

**`--no-fetch`**: passes the no-fetch flag through to every Status Collector call; no git fetch is performed anywhere in the run.

**`--non-default`**: filters the status table to only repositories currently on a non-default branch. The summary line and tips still reflect the filtered view.

**Exit codes**: exits 0 when all repositories are collected successfully. Exits non-zero when one or more repositories produce a `fetch-failed` error. `no-remote` does not cause a non-zero exit.

## Acceptance criteria

- [ ] `gcm status` with no flags shows a status table for all repositories under the clone root
- [ ] Header shows the clone root path: `Repos under <clone_root>:`
- [ ] Each row shows: relative path, branch, commits-behind count, dirty file count, applicable badges
- [ ] Sort order matches: non-default-branch first, then commits-behind descending, then alphabetical
- [ ] `--no-fetch` produces output without performing any git fetch (verified by absence of network activity in integration test with an unreachable remote)
- [ ] `--non-default` filters table to only non-default-branch repositories
- [ ] `[fetch-failed]` repositories appear in table; command exits non-zero
- [ ] `[no-remote]` repositories appear in table; command exits 0
- [ ] Summary line counts are correct
- [ ] Tips line is present with exact command strings
- [ ] Performance: completes in under 10 seconds for a 200-repository test fixture (measured in integration test)
- [ ] Integration tests use only local git repositories — no real network calls

## Blocked by

- 0003 — config store and config commands
- 0007 — status collector
- 0008 — status formatter
