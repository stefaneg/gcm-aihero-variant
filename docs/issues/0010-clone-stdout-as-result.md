# 0010 — `gcm clone` adopts stdout-as-result contract

Status: ready-for-agent

## What to build

Rework `gcm clone` so that stdout carries the result (the destination path) and stderr carries progress. This is the foundation that lets the shell-function wrapper (issue 0011) cd into the newly cloned working directory.

Behaviour:

- On a successful clone, the **only** thing on stdout is the destination path, emitted as a single bare line (no `S|` prefix, no surrounding decoration, no trailing summary). Progress messages — "Cloning to ...", completion notice — move to stderr.
- On a rerun against an already-present git repository whose `origin` URL matches the requested URL: re-emit the destination path on stdout, exit 0, stderr silent. No "Already cloned" line, no progress noise. Reruns look identical to fresh clones from the caller's perspective.
- Existing error paths (missing URL, non-git directory at destination, git clone failure) are unchanged in this slice except that they continue to produce **nothing** on stdout.

The bare-path stdout contract is load-bearing for shell integration and must not be decorated. Adding any prefix, suffix, or extra line on stdout will silently break the shell wrapper that ships in 0011.

The origin-mismatch case (destination is a git repo but origin URL differs) is handled in issue 0012; in this slice, fall through to the existing "destination exists but is not a git repository" code path for any existing-directory case other than matching-origin.

## Acceptance criteria

- [ ] `dest=$(gcm clone <url>)` captures the destination path verbatim with no extra characters
- [ ] Running `gcm clone <url>` interactively still shows progress to the user (on stderr)
- [ ] Idempotent rerun on a matching-origin repo prints only the destination path on stdout, nothing on stderr, exit 0
- [ ] All error paths emit nothing on stdout
- [ ] Existing clone integration tests updated to assert on stdout vs stderr separately
- [ ] New integration test: pipe `gcm clone` stdout to a file; the file contains exactly one line — the destination path

## Blocked by

- None
