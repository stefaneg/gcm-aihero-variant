# `git-clone-manager/internal/workerpool` — Ordered parallel map helper

## Glossary

No CONTEXT.md for this package: it exposes a generic concurrency primitive rather than domain vocabulary.

## Package scope

This package is a technical support primitive for running independent work items concurrently.

## Core concept owned

This package is the canonical home for ordered, error-preserving parallel mapping over a slice of items.

## Responsibilities

- Owns: the generic `Run` helper that processes items with a fixed worker count.
- Owns: `Result[T]`, the per-item value/error pair returned without aborting sibling work.
- Owns: preserving input order in the returned result slice.
- Does **not** own: repository status semantics; `internal/statuscollector` owns per-item work.
- Does **not** own: deciding whether any item error aborts a higher-level workflow; callers own aggregation.
- Does **not** own: user-facing progress, cancellation, or output; command and workflow packages own those concerns.

## Upstream (this package depends on)

- None.

## Downstream (consumers of this package)

- `git-clone-manager/internal/statuspipeline` — collects repository status concurrently while preserving order.

## Invariants & conventions

- Empty input returns an empty result slice without starting workers.
- Worker count is `2 * runtime.NumCPU()`.
- Each item writes exactly one slot, indexed by the original input position.
- Per-item errors are stored in results rather than stopping the pool.
- The work function must be safe for concurrent calls.

## When developing in this package

- [ ] Did any new concurrency feature preserve deterministic result ordering and leave cancellation/error policy to the
  caller?

## Clean-concept rating — PASS

This package owns a single concept cleanly: ordered parallel mapping. The smell test passes because the API is generic
and narrow, with no package-level mutable state and no domain or presentation leakage.
