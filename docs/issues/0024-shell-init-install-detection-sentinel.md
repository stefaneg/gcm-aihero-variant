# 0024 — shell-init: replace substring install detection with sentinel marker

Status: ready-for-agent

## What to build

`gcm shell-init --install` decides whether the wrapper is already present in the user's rc file via `strings.Contains(string(data), "gcm shell-init")`. This false-positives on any line mentioning the string — most obviously a user comment like `# disabled: gcm shell-init` or `# TODO re-enable gcm shell-init`. The install then silently skips and the user is confused why their wrapper isn't active.

Introduce a sentinel marker comment that the installer emits and detects. Something like `# gcm shell-init (managed) — do not edit this line` immediately above the install line. Detection becomes "does the rc file contain this exact marker line?" — robust to other mentions of the substring `gcm shell-init`.

Backwards compatibility: users who installed under the old detection have an rc file with the install line but no marker. Two acceptable approaches — implementer's choice:

1. Treat the absence of the marker as "not installed," let the installer append a fresh block (marker + line); the old line stays as a dead duplicate the user can clean up.
2. Detect either form: marker present *or* the legacy install-line pattern matches. New installs always write the marker.

Option (2) is friendlier; option (1) is simpler. Either is acceptable as long as a re-running user does not end up with two active wrappers in the same shell session.

## Acceptance criteria

- [ ] `gcm shell-init --install` writes a sentinel marker comment alongside the install line
- [ ] An rc file containing only a comment that mentions `gcm shell-init` (but no install) is detected as "not installed" and the installer proceeds
- [ ] An rc file containing the marker is detected as "installed" and the installer reports "Already installed in <path>"
- [ ] Re-running `gcm shell-init --install` on an rc file installed under the old detection does not produce two active wrappers (per the chosen backwards-compat strategy)
- [ ] Unit tests cover: fresh install, false-positive comment, idempotent re-run with marker present, the chosen backwards-compat case
- [ ] The marker text is stable enough to document (or kept as an unexported constant)

## Blocked by

- None - can start immediately
