# `git-clone-manager/internal/gitrunner` — Typed access to the `git` executable

## Glossary

Domain vocabulary lives in [`../../CONTEXT.md`](../../CONTEXT.md).

## Package scope

This package is the git integration boundary for repository operations used by clone and status workflows.

## Core concept owned

This package is the canonical home for running git commands and classifying git-process failures into typed errors.

## Responsibilities

- Owns: the `Runner` interface for clone, fetch, branch, dirty count, behind count, and default-branch queries.
- Owns: the concrete command runner backed by `os/exec`.
- Owns: typed error wrappers for missing git binaries, missing repositories, network failures, no remote, and unset
  `origin/HEAD`.
- Owns: fallback behaviour for behind-count calculation when `origin/HEAD` is unavailable.
- Does **not** own: repository discovery under the clone root; `internal/repositorywalker` owns filesystem walking.
- Does **not** own: turning git facts into status rows; `internal/statuscollector` owns status collection.
- Does **not** own: status table presentation; `internal/statusformatter` owns formatting and colour.

## Upstream (this package depends on)

- None.

## Downstream (consumers of this package)

- `git-clone-manager/internal/cmd` — creates a runner for clone and status command flows.
- `git-clone-manager/internal/gitrunnertest` — implements the runner interface for tests.
- `git-clone-manager/internal/statuscollector` — fetches and reads repository status facts through the runner.
- `git-clone-manager/internal/statuspipeline` — accepts the runner to wire collection across repositories.

## Invariants & conventions

- The only ambient process dependency is the configured git binary name.
- Repository-scoped commands are run with `git -C <repoPath>`.
- Stderr is preserved in `commandError` so classification can inspect git's human text while callers still unwrap.
- `DefaultBranch` requires `refs/remotes/origin/HEAD` to resolve to `origin/<branch>`.
- `CommitsBehind` falls back to `main` only when `origin/HEAD` is unset.

## When developing in this package

- [ ] Did any new git operation return a typed error that downstream collectors can distinguish from hard failures?

## See also

- [`../../CONTEXT.md`](../../CONTEXT.md) — default-branch, behind, dirty, and repository vocabulary.
- [`../../docs/adr/0003-protocol-passed-as-is.md`](../../docs/adr/0003-protocol-passed-as-is.md) — clone URL
  protocol is passed through to git.

## Clean-concept rating — PASS

This package owns a single concept cleanly: a typed boundary around the git executable. The smell test passes
because its several methods are facets of the same runner contract, with no package-level mutable state and no
self-admission comments.
