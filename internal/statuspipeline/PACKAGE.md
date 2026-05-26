# `git-clone-manager/internal/statuspipeline` — Repository status orchestration

## Glossary

Domain vocabulary lives in [`../../CONTEXT.md`](../../CONTEXT.md).

## Package scope

This package belongs to the status workflow's orchestration layer between discovery, per-repository collection, and
command presentation.

## Core concept owned

This package is the canonical home for collecting status results for every repository under a clone root.

## Responsibilities

- Owns: the `Pipeline` that combines repository walking, per-repository collection, and worker-pool execution.
- Owns: missing-clone-root policy for status, returning an empty result set when the root does not exist.
- Owns: converting worker-pool item errors into a single hard collection error.
- Does **not** own: filesystem discovery rules; `internal/repositorywalker` owns `.git` walking.
- Does **not** own: per-repository git-status semantics; `internal/statuscollector` owns `Result`.
- Does **not** own: concurrency primitives; `internal/workerpool` owns generic ordered parallelism.
- Does **not** own: status table formatting or command flags; `internal/statusformatter` and `internal/cmd` own them.

## Upstream (this package depends on)

- `git-clone-manager/internal/gitrunner` — supplies the runner passed into per-repository collectors.
- `git-clone-manager/internal/repositorywalker` — discovers repositories under the clone root.
- `git-clone-manager/internal/statuscollector` — collects one status result per repository.
- `git-clone-manager/internal/workerpool` — runs per-repository collection concurrently while preserving order.

## Downstream (consumers of this package)

- `git-clone-manager/internal/cmd` — executes the status pipeline for `gcm status`.

## Invariants & conventions

- A missing clone root is not a command failure for status; it returns zero collected repositories.
- Discovered repository order is preserved through worker-pool collection.
- Soft repository errors remain inside `statuscollector.Result`; hard errors abort the whole pipeline.
- The pipeline does not format output or inspect command flags beyond the `noFetch` boolean it receives.

## When developing in this package

- [ ] Did any orchestration change keep walking, collection, concurrency, and formatting in their owning packages?

## See also

- [`../../CONTEXT.md`](../../CONTEXT.md) — clone-root, repository, and status-table vocabulary.
- [ADR 0001](../../docs/adr/0001-filesystem-walk-for-repository-discovery.md) — status discovers repositories by
  walking the clone root.

## Clean-concept rating — PASS

This package owns a single concept cleanly: status collection orchestration over a clone root. The smell test passes
because it composes neighbouring packages without absorbing their policies, and it has no package-level mutable state
or self-admission comments.
