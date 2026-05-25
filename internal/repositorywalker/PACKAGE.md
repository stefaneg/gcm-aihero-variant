# `git-clone-manager/internal/repositorywalker` — Clone-root repository discovery

## Glossary

Domain vocabulary lives in [`../../CONTEXT.md`](../../CONTEXT.md).

## Package scope

This package belongs to the status workflow's discovery layer for repositories under the clone root.

## Core concept owned

This package is the canonical home for discovering repository directories by walking for `.git` directories.

## Responsibilities

- Owns: recursive clone-root traversal using `filepath.WalkDir`.
- Owns: identifying repository roots as parents of discovered `.git` directories.
- Owns: deterministic sorting of discovered repository paths.
- Does **not** own: handling a missing clone root as an empty status result; `internal/statuspipeline` owns that policy.
- Does **not** own: git status facts for discovered repositories; `internal/statuscollector` owns collection.
- Does **not** own: status table ordering; `internal/statusformatter` owns presentation sorting.

## Upstream (this package depends on)

- None.

## Downstream (consumers of this package)

- `git-clone-manager/internal/statuspipeline` — discovers repositories before parallel status collection.

## Invariants & conventions

- A `.git` directory marks its parent as a repository unless that parent is the clone root itself.
- Once a `.git` directory is found, the walker skips its contents.
- Results are sorted before returning so downstream concurrency does not control output order.
- Filesystem walk errors are returned unchanged for the pipeline to classify.

## When developing in this package

- [ ] Did any discovery change preserve the ADR rule that status is based on the filesystem, not a manifest?

## See also

- [`../../CONTEXT.md`](../../CONTEXT.md) — repository and clone-root vocabulary.
- [ADR 0001](../../docs/adr/0001-filesystem-walk-for-repository-discovery.md) — decision to discover repositories
  by walking the filesystem.

## Clean-concept rating — PASS

This package owns a single concept cleanly: filesystem discovery of repository roots. The smell test passes because
the package has one exported function, no package-level mutable state, and no unrelated presentation or git-process
concerns.
