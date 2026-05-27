# 0018 — statuscollector: stop guessing default branch; use `errors.As` consistently

Status: ready-for-agent

## What to build

Two correctness fixes in `internal/statuscollector` that touch the same error-handling section.

**Stop guessing the default branch when no remote exists.** When the git runner returns `NoRemoteError`, the collector currently falls back to `defaultBranch = "main"`. This directly contradicts CONTEXT.md ("when the repository has no remote, the concept does not apply") and PRD story #25 ("gcm never lies about a guessed default branch"). The lie is presently masked by the formatter (which skips the `[!main]` badge when `ErrorState == NoRemote`), but `Result.DefaultBranch` is a public field and will leak the fake value to any future consumer — `gcm pull`, JSON output, new badges, downstream tests. Leave `DefaultBranch` as the zero value (empty string) on `NoRemoteError`. The formatter and any callers must treat empty `DefaultBranch` as "not applicable", consistent with the no-remote ErrorState.

The PRD's "Testing Decisions" section still mentions the `"main"` fallback test as a residue of the old behaviour — remove that residue.

**Use `errors.As` consistently in the same file.** The error classification block uses a concrete type switch (`switch err.(type)`) in one place and `errors.As` in another. A future wrapped error (`fmt.Errorf("...: %w", &NetworkError{...})`) would silently fall through the type switch to `default` and abort the row. Convert the type switch to `errors.As` to match the rest of the file.

## Acceptance criteria

- [ ] A repository with no remote produces a `Result` whose `DefaultBranch` is empty and whose `ErrorState` indicates no-remote
- [ ] The status table for a no-remote repository renders identically to today (no `[!main]` badge, `[no-remote]` badge present)
- [ ] A wrapped `NetworkError` (`fmt.Errorf("...: %w", &NetworkError{...})`) is classified the same as the bare error
- [ ] All existing `statuscollector` tests pass; tests asserting `DefaultBranch == "main"` for no-remote repos are updated to assert empty
- [ ] The residual `"main"` fallback test mention in the PRD testing section is removed
- [ ] No new external callers of `Result.DefaultBranch` are introduced

## Blocked by

- None - can start immediately
