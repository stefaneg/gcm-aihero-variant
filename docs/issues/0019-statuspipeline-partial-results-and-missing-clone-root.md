# 0019 — statuspipeline: partial results on hard error; actionable error when clone root missing

Status: ready-for-agent

## What to build

Two related visibility fixes in `internal/statuspipeline` so users always see either the table or a useful error.

**Partial results on hard per-repository errors.** Today the pipeline aborts the entire batch on the first per-repo error that wasn't classified as soft (network, no-remote, origin-HEAD-unset). A corrupt `.git`, a permission failure, or any `exec.ExitError` whose stderr doesn't match a known keyword therefore wipes out every other row. PRD story #40 requires the opposite: "partial batch failures produce a non-zero exit *with a partial result shown*." Convert hard per-item errors into a result row with an error state (e.g. `ErrorStateUnknown` or reuse `ErrorStateFetchFailed` if semantically appropriate — implementer's call as long as the row renders with an explanatory badge) and record a batch-level error to propagate the non-zero exit code separately.

**Actionable error when the clone root does not exist.** Today the pipeline silently returns `(nil, nil)` when the clone root directory is absent, so `gcm status` renders `0 repos — 0 current, 0 behind, 0 non-default-branch` with no hint that the configured clone root is missing. This violates PRD story #38 ("state what happened, why, and what to do next"). `cmd/open.go` already handles the same case correctly with a distinct error pointing at `gcm config set clone-root`; mirror that pattern here. The "exists but empty" case is unchanged and continues to render an empty table.

## Acceptance criteria

- [ ] A batch in which one repository returns a hard error (e.g. corrupt `.git`) still renders all other rows; the failing row shows an error badge; `gcm status` exits non-zero
- [ ] The hard-error row's badge is greppable text (per PRD #19), consistent with existing badges
- [ ] `gcm status` against a configured-but-nonexistent clone root exits non-zero with an actionable message that names the path and points at `gcm config set clone-root`
- [ ] `gcm status` against an existing-but-empty clone root behaves exactly as today (renders the empty-table form, exit 0)
- [ ] Unit tests cover: mixed-success batch with one hard-error row, missing clone root error, empty clone root unchanged
- [ ] Existing pipeline tests pass

## Blocked by

- None - can start immediately
