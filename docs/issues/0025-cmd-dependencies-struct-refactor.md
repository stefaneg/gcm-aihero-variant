# 0025 — internal/cmd: replace mutable package-level test seams with a `Dependencies` struct

Status: done

## What to build

`internal/cmd` uses package-level mutable function variables as test seams: `loadEffectiveCloneConfig`, `newGitRunner`, `loadEffectiveStatusConfig`, `newStatusCollector`, `runOpenFZF`, `openFZFAvailable`, `openWriterIsTTY`, `loadEffectiveShellConfig`, and more added incrementally. The package's own `PACKAGE.md` self-rates this 6/10 and flags it as a known liability.

The consequence is that no test in the package can call `t.Parallel()` — every test reaches into globals and would race with siblings. As the package grows (0015, 0016, 0019, 0021, 0022 all add code here) the seam list grows with it.

Replace the globals with a `Dependencies` struct (or similar — `CommandDeps`, `Runtime`, implementer's choice) that holds the same function-typed fields. Construct a default instance in `NewRootCommand` (or wherever the root is built) and thread it through to each subcommand constructor. Tests construct their own `Dependencies` with stubs and pass it to the command-under-test. No more globals; tests can parallelise.

Keep the migration mechanical: this is a refactor, not a redesign. Same fields, same signatures, just moved into a struct and passed explicitly. Update the package's `PACKAGE.md` self-assessment when done.

There is design surface here (struct name, where it lives, whether per-subcommand `Dependencies` or one shared struct) but it's small enough to resolve in the PR review rather than upfront. If the implementer hits a genuine design fork they can flag it as a comment for human review.

## Acceptance criteria

- [ ] All previously-mutable package-level test-seam variables in `internal/cmd` are removed or made unexported and immutable
- [ ] Each subcommand constructor takes its dependencies as parameters (directly or via a struct)
- [ ] `NewRootCommand` (or equivalent) constructs the production `Dependencies` and threads it to every subcommand
- [ ] At least one test file in `internal/cmd` calls `t.Parallel()` and passes (proof the seam removal succeeded)
- [ ] No production behaviour change — all existing tests still pass without modification beyond construction-style changes
- [ ] The `PACKAGE.md` self-assessment is updated to reflect the new structure

## Blocked by

- None - can start immediately (last-to-merge will rebase against any cmd/ changes from 0019/0021/0022)
