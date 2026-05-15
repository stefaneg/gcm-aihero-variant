# 0006 — Repository walker and worker pool

## What to build

Two foundational modules used by `gcm status` (and post-v1 commands).

**Repository Walker**: walks the clone root directory tree and returns the path of every directory that contains a `.git` subdirectory. All directories found are returned regardless of whether gcm created them. No git operations — filesystem traversal only. Follows symlinks is not required for v1.

**Worker Pool**: a generic parallel executor sized at 2× the number of available CPU cores. Accepts a slice of work items of any type and a function to apply to each item; returns a slice of results in input order with per-item errors preserved. The pool has no knowledge of repositories, git, or status — it is a pure concurrency primitive.

## Acceptance criteria

- [ ] Repository Walker returns all directories with a `.git` subdirectory under the clone root
- [ ] Repository Walker returns an empty slice (not an error) when the clone root is empty
- [ ] Repository Walker returns an error if the clone root does not exist
- [ ] Repository Walker does not return the clone root itself if it happens to be a git repository
- [ ] Worker Pool runs work items in parallel using 2× `runtime.NumCPU()` goroutines
- [ ] Worker Pool preserves input order in the results slice
- [ ] Worker Pool propagates per-item errors without aborting other items
- [ ] Worker Pool unit tests verify parallelism by timing concurrent work items that each sleep briefly
- [ ] Repository Walker unit tests use a temp directory tree

## Blocked by

- 0001 — project scaffold and CLI skeleton
