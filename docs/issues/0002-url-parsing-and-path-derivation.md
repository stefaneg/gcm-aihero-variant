# 0002 — URL parsing and path derivation

## What to build

Two pure, tested modules with no I/O:

**URL Parser**: accepts a git URL in any common format (https://, ssh://, git@) and returns three components: hostname, path prefix, and repository name. Strips trailing `.git` suffix. Handles missing path prefix (bare `host/repo` with no intermediate segments). Returns a typed error for malformed input.

**Path Deriver**: accepts a clone root, hostname, path prefix, and repository name and returns the derived path: `${clone_root}/${hostname}/${path_prefix}/${repo_name}`. When path prefix is empty the result is `${clone_root}/${hostname}/${repo_name}`.

These two modules encode the core invariant of gcm: the derived path is deterministic and mirrors the full remote URL hierarchy. The algorithm must not change without a versioned migration path, as it is the discovery key for `gcm status`.

## Acceptance criteria

- [ ] URL Parser handles `https://github.com/nWave-ai/nWave` → hostname: `github.com`, path prefix: `nWave-ai`, repo name: `nWave`
- [ ] URL Parser handles `git@github.com:nWave-ai/nWave.git` → same components, `.git` stripped
- [ ] URL Parser handles `ssh://git@github.com/nWave-ai/nWave` → same components
- [ ] URL Parser handles deep path prefix: `https://gitlab.com/group/subgroup/repo` → path prefix: `group/subgroup`, repo name: `repo`
- [ ] URL Parser handles bare `https://example.com/repo` (no path prefix) → path prefix: `""`, repo name: `repo`
- [ ] URL Parser returns a typed error for malformed URLs
- [ ] Path Deriver produces `~/src/github.com/nWave-ai/nWave` given clone root `~/src` and parsed components from the above
- [ ] Path Deriver handles empty path prefix correctly (no double slash)
- [ ] Unit test table covers all URL formats and edge cases above

## Blocked by

- 0001 — project scaffold and CLI skeleton
