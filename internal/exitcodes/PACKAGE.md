# `git-clone-manager/internal/exitcodes` — Exit-code-bearing errors

## Glossary

No CONTEXT.md for this package: it exposes technical process primitives rather than internal domain concepts.

## Package scope

This package is a small CLI support kernel shared by command execution and tests that inspect process results.

## Core concept owned

This package is the canonical home for mapping Go errors to `gcm` process exit codes.

## Responsibilities

- Owns: the numeric exit-code constants for success, general failure, and usage failure.
- Owns: the `Error` wrapper that carries an exit code while preserving error unwrapping.
- Owns: helpers for creating usage errors and extracting a process code from any error.
- Does **not** own: deciding which command failures are usage failures; `internal/cmd` owns command classification.
- Does **not** own: printing errors or terminating the process; `internal/cmd` and `cmd/gcm` own those boundaries.
- Does **not** own: domain-specific error types from git or repository parsing; their packages own those errors.

## Upstream (this package depends on)

- None.

## Downstream (consumers of this package)

- `git-clone-manager/internal/cmd` — wraps Cobra validation errors and returns process codes from `Execute`.
- `git-clone-manager/cmd/gcm` — checks process codes in executable-level tests.

## Invariants & conventions

- `nil` errors always map to `Success`.
- Any error implementing `ExitCode() int` controls its own code.
- Errors without an exit-code interface map to `General`.
- `WithCode` preserves `nil` as `nil`, so callers can wrap optional failures without introducing an error.

## When developing in this package

- [ ] Did any new code preserve `errors.As` compatibility by adding behaviour through interfaces rather than
  concrete type checks?

## Clean-concept rating — PASS

This package owns a single concept cleanly: translating errors into CLI exit codes. The smell test passes because
the surface is tiny, there is no package-level mutable state, and all responsibilities revolve around one wrapper
contract.
