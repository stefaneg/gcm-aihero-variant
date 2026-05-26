# `git-clone-manager/internal/gitrunnertest` — Thread-safe fake git runner for tests

## Glossary

No CONTEXT.md for this package: it provides test doubles for an external command boundary rather than domain
vocabulary.

## Package scope

This package is test support for packages that depend on `internal/gitrunner.Runner`.

## Core concept owned

This package is the canonical home for a configurable, call-recording fake implementation of `gitrunner.Runner`.

## Responsibilities

- Owns: `Fake`, a mutex-protected `gitrunner.Runner` test double.
- Owns: per-method stubs for clone, fetch, branch, dirty-count, behind-count, and default-branch behaviour.
- Owns: call recording accessors for assertions in command and collector tests.
- Does **not** own: real git command behaviour or error classification; `internal/gitrunner` owns that boundary.
- Does **not** own: package-specific test assertions; each consuming test package owns expected behaviour.
- Does **not** own: production status collection semantics; `internal/statuscollector` owns those.

## Upstream (this package depends on)

- `git-clone-manager/internal/gitrunner` — provides the interface the fake implements.

## Downstream (consumers of this package)

- `git-clone-manager/internal/cmd` tests — inject fake clone behaviour and assert clone calls.
- `git-clone-manager/internal/statuscollector` tests — stub git facts and error cases during collection.

## Invariants & conventions

- Every method that reads or writes fake state must hold `mu`.
- Accessors return defensive copies so assertions cannot mutate recorded calls.
- A nil stub means "return the configured value and nil error" for read methods, and nil for side-effect methods.
- The package must remain production-buildable even though its only consumers are tests.

## When developing in this package

- [ ] Did every new fake capability record calls and expose assertions without sharing mutable slices with tests?

## Clean-concept rating — PASS

This package owns a single concept cleanly: a thread-safe fake for the git runner boundary. The smell test passes
because all mutable state is instance-scoped behind a mutex, with no package-level mutable state or self-admission
comments.
