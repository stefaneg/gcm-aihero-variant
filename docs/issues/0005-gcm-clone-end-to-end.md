# 0005 — `gcm clone` end-to-end

## What to build

The complete `gcm clone <url>` command wired through all layers: URL Parser → Path Deriver → Config Store (for clone root) → Git Runner (for the actual clone).

Behaviour:
- Parse the URL and derive the local path; display the derived path before any cloning begins
- If the clone root does not exist, warn the user and create it, then proceed
- If the destination already exists and is a git repository, print "Already cloned at <path>" and exit 0
- If the destination exists but is not a git repository, print an actionable error (what / why / what to do) and exit 1
- Otherwise clone and confirm the path

The URL is passed to git as-is — no scheme rewriting. Protocol is whatever is in the URL.

## Acceptance criteria

- [ ] `gcm clone https://github.com/nWave-ai/nWave` clones to `~/src/github.com/nWave-ai/nWave` (or configured clone root)
- [ ] Derived path is printed before cloning starts
- [ ] Clone root is created with a warning if it does not exist
- [ ] Re-running the same clone prints "Already cloned at <path>" and exits 0 without modifying the repository
- [ ] Destination exists as a non-git directory → actionable error, exit 1
- [ ] Missing URL argument → usage error, exit 2
- [ ] URL with no matching scheme falls through to git unchanged
- [ ] Cloned repository appears in `gcm status` output immediately after (verified in integration test)
- [ ] Integration tests use a local bare repository as the remote — no real network calls

## Blocked by

- 0002 — URL parsing and path derivation
- 0003 — config store and config commands
- 0004 — git runner
