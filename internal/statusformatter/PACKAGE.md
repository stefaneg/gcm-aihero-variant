# `git-clone-manager/internal/statusformatter` — Plain-text status table rendering

## Glossary

Domain vocabulary lives in [`../../CONTEXT.md`](../../CONTEXT.md).

## Package scope

This package belongs to the status workflow's presentation layer.

## Core concept owned

This package is the canonical home for rendering collected repository status results as the `gcm status` table.

## Responsibilities

- Owns: table row ordering for non-default branches, behind counts, and derived path.
- Owns: clone-root-relative path rendering.
- Owns: summary counts for current, behind, and non-default repositories.
- Owns: status badges and optional ANSI colour selection.
- Does **not** own: collecting git facts; `internal/statuscollector` owns the `Result` values.
- Does **not** own: deciding whether stdout is a TTY or whether `NO_COLOR` is set; `internal/cmd` supplies options.
- Does **not** own: command flags or process exit behaviour; `internal/cmd` owns CLI flow.

## Upstream (this package depends on)

- `git-clone-manager/internal/statuscollector` — provides result rows and soft-error states.

## Downstream (consumers of this package)

- `git-clone-manager/internal/cmd` — formats the collected results for `gcm status`.

## Invariants & conventions

- Paths are rendered relative to the clone root and returned as an error if that relationship cannot be computed.
- Colour appears only when stdout is a TTY and `NoColor` is false.
- Soft-error rows are excluded from current and behind summary counts unless they are non-default.
- The output always includes the header, one row per result, a summary line, and the tips line.
- The formatter returns a string and does not write to stdout or stderr.

## When developing in this package

- [ ] Did any presentation change keep status vocabulary aligned with `CONTEXT.md` and leave data collection outside
  the formatter?

## See also

- [`../../CONTEXT.md`](../../CONTEXT.md) — status-table, behind, dirty, and non-default-branch vocabulary.

## Clean-concept rating — PASS

This package owns a single concept cleanly: rendering the status table. The smell test passes because sorting,
badges, colour, and summary text are all presentation facets over `statuscollector.Result`, with no package-level
mutable state.
