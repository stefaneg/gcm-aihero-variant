# `git-clone-manager/internal/statuscollector` — Per-repository status facts

## Glossary

Domain vocabulary lives in [`../../CONTEXT.md`](../../CONTEXT.md).

## Package scope

This package belongs to the status workflow's per-repository collection layer.

## Core concept owned

This package is the canonical home for turning git runner facts into one repository status result.

## Responsibilities

- Owns: `Result`, the structured status facts for one repository.
- Owns: `ErrorState`, the soft-error vocabulary that can appear in a status table row.
- Owns: `Collector`, which fetches when requested and reads branch, default branch, behind, and dirty facts.
- Owns: no-remote and fetch-failed downgrade policy for soft errors.
- Does **not** own: discovering which repositories to collect; `internal/repositorywalker` and
  `internal/statuspipeline` own that.
- Does **not** own: executing git commands directly; `internal/gitrunner` owns command execution and typed errors.
- Does **not** own: row formatting, colour, or sorting for display; `internal/statusformatter` owns presentation.

## Upstream (this package depends on)

- `git-clone-manager/internal/gitrunner` — provides repository facts and typed git errors.

## Downstream (consumers of this package)

- `git-clone-manager/internal/cmd` — filters results for `--non-default` and checks fetch-failed exit behaviour.
- `git-clone-manager/internal/statusformatter` — formats `Result` rows and `ErrorState` badges.
- `git-clone-manager/internal/statuspipeline` — collects one `Result` for each discovered repository.

## Invariants & conventions

- Network fetch failures become `ErrorStateFetchFailed`; no-remote failures become `ErrorStateNoRemote`.
- Unexpected git errors abort collection by returning an error.
- When `origin/HEAD` is unset or no remote exists, the default branch is left empty and surfaced through an error state.
- No-remote repositories skip behind-count calculation but still report dirty count.
- The collector is safe to use concurrently if the injected runner is safe.

## When developing in this package

- [ ] Did any new soft-failure case become an `ErrorState` that the formatter and command exit behaviour can handle?

## See also

- [`../../CONTEXT.md`](../../CONTEXT.md) — dirty, behind, default-branch, and status-table vocabulary.

## Clean-concept rating — PASS

This package owns a single concept cleanly: collecting one repository's structured status facts. The smell test
passes because its policies all converge on `Result`, with no package-level mutable state and no self-admission
comments.
