# `git-clone-manager/internal/cmd` — Cobra command assembly and CLI workflow execution

## Glossary

Domain vocabulary lives in [`../../CONTEXT.md`](../../CONTEXT.md).

## Package scope

This package is the CLI application layer for wiring commands to configuration, URL parsing, git execution, status
collection, and status formatting.

## Core concept owned

This package is the canonical home for translating `gcm` command invocations into package-level application workflows.

## Responsibilities

- Owns: root command construction and Cobra subcommand registration.
- Owns: `Execute`, including writer injection, error printing, usage-error normalisation, and exit-code selection.
- Owns: clone command flow from config lookup through derived path inspection and git clone execution.
- Owns: config command flow for setting and showing `clone_root`.
- Owns: status command flow for flags, TTY/no-colour option detection, result filtering, output writing, and
  fetch-failure exit behaviour.
- Does **not** own: config file parsing or defaults; `internal/configstore` owns them.
- Does **not** own: repository URL parsing or derived-path computation; `internal/repourl` owns them.
- Does **not** own: git command execution and typed git errors; `internal/gitrunner` owns them.
- Does **not** own: repository discovery and per-repository status collection; `internal/statuspipeline` owns the
  orchestration.
- Does **not** own: plain-text status table rendering; `internal/statusformatter` owns presentation layout.
- Does **not** own: exit-code wrapper semantics; `internal/exitcodes` owns the error-to-code contract.

## Upstream (this package depends on)

- `git-clone-manager/internal/configstore` — loads and saves the clone-root configuration.
- `git-clone-manager/internal/exitcodes` — marks usage errors and maps command errors to process codes.
- `git-clone-manager/internal/gitrunner` — clones repositories and creates the runner used by status.
- `git-clone-manager/internal/repourl` — parses clone URLs and derives destination paths.
- `git-clone-manager/internal/statuscollector` — supplies status result fields and soft-error states for filtering.
- `git-clone-manager/internal/statusformatter` — renders the status table.
- `git-clone-manager/internal/statuspipeline` — collects repository status results under the clone root.

## Downstream (consumers of this package)

- `git-clone-manager/cmd/gcm` — runs `Execute` from the executable entry point.

## Invariants & conventions

- Commands must use injected Cobra stdout/stderr writers rather than ambient stdout/stderr.
- Cobra usage and flag errors are wrapped as usage errors so they exit with code 2.
- `Execute` prints exactly one normalised error line to stderr when command execution fails.
- `clone` passes the raw URL to git and uses `repourl` only for the destination path.
- Existing git repositories at the destination are idempotent; non-git destination paths block cloning.
- `status --non-default` still treats filtered fetch failures as command failures.

## When developing in this package

- [ ] Did the command change keep domain work in the owning internal package and route all user-visible output through
  Cobra's injected writers?

## See also

- [`../../CONTEXT.md`](../../CONTEXT.md) — command-facing clone-root, derived-path, and status-table vocabulary.
- [`../../docs/adr/0002-zero-config-clone.md`](../../docs/adr/0002-zero-config-clone.md) — clone should run without
  prior configuration.
- [`../../docs/adr/0003-protocol-passed-as-is.md`](../../docs/adr/0003-protocol-passed-as-is.md) — clone passes the
  supplied URL to git unchanged.

## Clean-concept rating — 6/10

A reviewer assessment of how cleanly this package owns a single concept. Captured here so the next reader sees the
known design debt up front; revisit when any of the items below is addressed.

The package is a coherent CLI application layer, but it currently mixes command assembly, filesystem preparation,
presentation-adjacent option detection, and test seams. That is acceptable at the current size, but the package
already fails the smell test through package-level mutable state and several distinct workflow concerns.

**Strengths**

- `Execute` in `execute.go` cleanly centralises writer injection, error normalisation, and exit-code selection.
- `NewRootCommand` in `root.go` keeps Cobra assembly explicit and routes usage errors through `exitcodes`.
- `status.go` delegates data collection and formatting to `statuspipeline` and `statusformatter`.

**What costs the points**

1. `clone.go` lines 18-23 keep `loadEffectiveCloneConfig` and `newGitRunner` as package-level mutable function
   variables; move these behind a command dependency struct or constructor parameter.
2. `clone.go` combines home expansion, clone-root creation, destination inspection, and command messaging in the same
   file as Cobra wiring; split filesystem preparation behind a small command-local service if clone grows.
3. `status.go` owns TTY and `NO_COLOR` detection alongside flag parsing and exit policy; keep it here only while it
   remains command wiring, or move option detection to a presentation boundary if more output modes appear.

**Path to 7**: Replace package-level mutable test seams with explicit command dependencies, then keep clone filesystem
helpers private to a narrow command service so the package remains an application layer rather than a utility sink.
