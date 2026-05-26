# `git-clone-manager/internal/configstore` — User configuration storage and defaults

## Glossary

Domain vocabulary lives in [`../../CONTEXT.md`](../../CONTEXT.md).

## Package scope

This package belongs to the configuration support layer for the single-context `gcm` CLI.

## Core concept owned

This package is the canonical home for loading, defaulting, parsing, and saving the effective `gcm`
configuration.

## Responsibilities

- Owns: the `Store` abstraction for locating configuration via `GCM_CONFIG` or the user home directory.
- Owns: the `Config` and `EffectiveConfig` shapes for persisted and runtime configuration.
- Owns: the default clone root value and the minimal `clone_root` YAML-compatible scalar format.
- Does **not** own: command text or user-facing command flow; `internal/cmd` owns CLI presentation.
- Does **not** own: expanding `~` in clone roots for filesystem use; `internal/cmd` owns command input
  normalisation.
- Does **not** own: repository URL-derived paths under the clone root; `internal/repourl` owns that mapping.

## Upstream (this package depends on)

- None.

## Downstream (consumers of this package)

- `git-clone-manager/internal/cmd` — reads effective config for `clone` and `status`, and writes
  `clone_root` from `config set clone-root`.

## Invariants & conventions

- Missing config files are not errors; they yield `EffectiveConfig{CloneRoot: "~/src", CloneRootIsDefault: true}`.
- `GCM_CONFIG` is an absolute override for the config file path and bypasses the default home-relative path.
- Writes create the parent directory and force the config file mode to `0600`.
- The parser intentionally recognises only `clone_root`; unknown keys are ignored rather than rejected.
- The package returns contextual Go errors and does not print or exit.

## When developing in this package

- [ ] Did any new persisted setting preserve zero-config behaviour when the file is absent or only partially
  populated?

## See also

- [`../../CONTEXT.md`](../../CONTEXT.md) — clone-root vocabulary and relationships.
- [`../../docs/adr/0002-zero-config-clone.md`](../../docs/adr/0002-zero-config-clone.md) — zero-config clone
  decision that the default config path supports.

## Clean-concept rating — PASS

This package owns a single concept cleanly: the effective configuration file contract. The smell test passes
because the package has one exported store surface, no package-level mutable state, and no self-admission comments.
