# 0022 — gcm clone: actionable error for half-broken `.git`; race-free `mkdirAllTracked`

Status: ready-for-agent

## What to build

Two fixes in `internal/cmd/clone.go`'s destination-validation and -creation paths.

**Actionable error for a half-broken `.git`.** When the destination contains a `.git` entry, the code calls `runner.OriginURL` to read the origin URL. If the `.git` is corrupt (a half-completed clone aborted out-of-band, a partial copy, etc.) `OriginURL` returns a low-level error that bubbles up as `"inspect destination origin <path>: ..."`. PRD story #7 requires an actionable error here: "destination exists but is not a git repository (or is a broken one); move or remove it first, then run gcm clone again." Detect the `OriginURL` failure mode and return the actionable form. The healthy case (`.git` exists, `OriginURL` returns the URL we expected) is unchanged. The URL-mismatch case (PRD #45) is unchanged.

**Race-free `mkdirAllTracked`.** The current implementation calls `os.Stat(current)` and, on `ErrNotExist`, plans to create `current` and record it for cleanup. Between the `Stat` and the eventual `os.MkdirAll`, another process can create `current`; the cleanup loop will then `os.Remove` a directory the user didn't ask gcm to manage. (The loop uses non-recursive `os.Remove`, which limits the blast radius, but it's still wrong.) Replace the stat-then-mkdir-all pattern with per-segment `os.Mkdir`: walk the path components, attempt `os.Mkdir` on each missing one, and only record a segment as "created by gcm" when the `Mkdir` syscall itself returned nil (i.e. *we* created it, not someone else). `os.IsExist` on a segment means we leave it alone — including in cleanup.

The cleanup loop's existing semantics ("remove only what gcm created, leaf-first, stop on any non-empty error") stay the same; the input set just becomes authoritative.

## Acceptance criteria

- [ ] A destination containing a corrupt `.git` directory produces the actionable PRD #7 error, not "inspect destination origin: ..."
- [ ] A destination containing a valid `.git` with matching origin URL clones as today (idempotent rerun)
- [ ] A destination containing a valid `.git` with a different origin URL produces the PRD #45 error as today
- [ ] `mkdirAllTracked` records a directory in the cleanup set if and only if the per-segment `os.Mkdir` for that directory returned nil
- [ ] If a concurrent process creates an intermediate directory between gcm's segment-by-segment walk, gcm leaves that directory alone on both success and cleanup paths
- [ ] Unit tests cover: corrupt-`.git` actionable error; race scenario (intermediate dir created between segments — can be simulated with a fake filesystem or by pre-creating directories with `os.Mkdir` before the call)
- [ ] Existing clone tests pass

## Blocked by

- None - can start immediately
