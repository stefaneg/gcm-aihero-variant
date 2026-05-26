# CLI Surface: v1 Core Commands

Generated: 2026-05-25T13:27Z
Last regenerated: 2026-05-25T13:27Z

## 0. Source artifacts

- Input: `docs/prd/v1-core-commands.md`
- Binary: `gcm` (entry point: `cmd/gcm/main.go`; command tree: `internal/cmd/`)
- Conventions consulted:
  - `docs/specs/output-spec.md`
  - `docs/specs/gcm-cmd-output-spec.md`: absent - author as follow-up only if per-command elaboration becomes necessary
  - existing Cobra tree under `internal/cmd/`
  - existing exit-code package: `internal/exitcodes`
  - existing status formatter: `internal/statusformatter`
- Hint system: optional/absent. `docs/specs/output-spec.md` explicitly says there is no hint system, so errors use one-part `Error:` lines and screens do not render `Hint:` lines.

## 1. Output variants

| # | Variant | Type | Surface elements implied |
|---|---|---|---|
| V1 | Clone repository - new clone success | happy path | `gcm clone <url>`, derived destination line, root creation warning, completion line |
| V2 | Clone repository - already present | idempotent happy path | existing git repo detection, skip message, exit 0 |
| V3 | Clone repository - destination blocked | error path | non-git destination conflict, exit 1, one-part error |
| V4 | Clone repository - invalid or missing URL | error path | URL positional validation, exit 2 for missing, exit 1 for malformed |
| V5 | Status - default table | happy path | `gcm status`, fetch by default, table, badges, summary, tips |
| V6 | Status - no fetch | happy path | `--no-fetch`, local-state table |
| V7 | Status - non-default filter | happy path | `--non-default`, filtered table |
| V8 | Status - partial fetch failure | partial failure | table with `[fetch-failed]`, stderr error, exit 1 |
| V9 | Status - no repositories | happy path | empty table state, summary, exit 0 |
| V10 | Config set clone-root - success | state change | `gcm config set clone-root <path>`, config path result |
| V11 | Config set clone-root - missing path | error path | required positional, exit 2 |
| V12 | Config show - default plain | happy path | effective config with default annotation |
| V13 | Discoverability - root and group help | discoverability | `gcm`, `gcm config`, `gcm help <command>`, completion command |
| V14 | Unknown command or flag | usage error | Cobra-style usage error normalized to exit 2 |

The PRD stories map directly: clone stories 1-8 to V1-V4, status stories 9-33 and 40 to V5-V9, config stories 34-37 to V10-V12, and error/exit stories 38-39 to V3, V4, V8, V11, and V14.

## 2. Primary human workflow

Most common user goal: A developer clones a repository by URL and immediately knows the exact local path where it landed.

Shortest happy-path invocation:

```console
$ gcm clone git@github.com:nWave-ai/aihero.git
```

Input acquisition:

| Input | Strategy | Notes |
|---|---|---|
| repository URL | positional (arg 1) | required; passed to git as-is |
| clone root | env-var | `GCM_CONFIG` can point to config file; otherwise effective `clone_root` defaults to `~/src` |
| destination path | n/a | derived from clone root and parsed URL hierarchy |
| confirmation | n/a | clone is idempotent; existing git repo is skipped |

First success line:

```text
S| Cloning to /Users/gulli/src/github.com/nWave-ai/aihero...
```

First failure line for the most likely error:

```text
E| Error: destination "/Users/gulli/src/github.com/nWave-ai/aihero": exists but is not a git repository. Move or remove it, then run gcm clone again.
```

## 3. Command tree

Layout: Existing tree is verb-first at the root for actions (`clone`, `status`) and noun-first under `config`. Keep that shape; no existing commands are renamed or moved.

Subcommands:

| Path | Purpose | Variants covered |
|---|---|---|
| `gcm clone <url>` | Clone a repository into its derived path | V1, V2, V3, V4 |
| `gcm status` | Show repository status under the clone root | V5, V6, V7, V8, V9 |
| `gcm config set clone-root <path>` | Persist the clone root override | V10, V11 |
| `gcm config show` | Show effective configuration, including defaults | V12 |
| `gcm help [command]` | Show help for root or a subcommand | V13 |
| `gcm completion bash|zsh|fish|powershell` | Generate shell completion scripts | V13 |

Aliases: none. The command names are already short and script-friendly.

Tree diff against current `gcm --help`: no command additions, renames, or moves. The v1 surface keeps the existing root commands `clone`, `status`, `config`, `help`, and `completion`.

## 4. Flag set per subcommand

Global flags:

| Long | Short | Type | Default | Required | Env var | Mutex group | Notes |
|---|---|---|---|---|---|---|---|
| `--help` | `-h` | bool | `false` | no |  |  | provided by Cobra |

`gcm clone <url>`: no command-specific flags.

`gcm status`:

| Long | Short | Type | Default | Required | Env var | Mutex group | Notes |
|---|---|---|---|---|---|---|---|
| `--no-fetch` |  | bool | `false` | no |  |  | use local git state without fetching remotes first |
| `--non-default` |  | bool | `false` | no |  |  | show only repositories on non-default branches |

`gcm config set clone-root <path>`: no command-specific flags.

`gcm config show`: no command-specific flags.

Short flags: only Cobra's built-in `-h`. No new short flags are added in v1.

The surface deliberately excludes `--format`, `--output`, `--dry-run`, `--color`, and `--no-color` per `docs/specs/output-spec.md`.

## 5. Output schema and stream contract

Output mode:

| Subcommand | Default-format result | Rationale | Structured schema ref | Side-effect-only? |
|---|---|---|---|---|
| `gcm clone` | result lines with destination path and completion | the destination is the value the human needs to verify before and after git runs | n/a - no JSON output | no |
| `gcm status` | aligned table with summary and tips | many repositories are fastest to scan as rows with stable columns and text badges | n/a - no JSON output | no |
| `gcm config set clone-root` | one-line saved-path confirmation | confirms the write and names the config file touched | n/a - no JSON output | no |
| `gcm config show` | key/value lines with default annotation | config has few fields; labelled lines are easier than a table | n/a - no JSON output | no |

TTY-aware behavior:

| Subcommand | TTY-aware behavior |
|---|---|
| `gcm clone` | no color; lines are unchanged when piped |
| `gcm status` | badge text is always emitted; badge color is added only when stdout is a TTY and `NO_COLOR` is unset |
| `gcm config set clone-root` | no color; lines are unchanged when piped |
| `gcm config show` | no color; lines are unchanged when piped |

Output-spec compliance for this command set:

- stdout vs stderr: result data goes to stdout; diagnostics, root-creation warnings, and errors go to stderr.
- side-effect-only: none of the v1 commands are silent side-effect-only commands.
- structured output: no JSON or alternate output format is part of the v1 contract.

## 6. Exit-code map

| Code | Class | Trigger | stderr substring (canonical) |
|---|---|---|---|
| 0 | success | normal completion, including idempotent clone skip and `no-remote` status rows | - |
| 1 | general failure | git clone failure, unreadable clone root, blocked destination, malformed URL, status collection failure, or partial fetch failure | starts with `Error: ...` |
| 2 | usage/config error | missing argument, unknown command, unknown flag, config read/write/parse failure | starts with `Error: ...` |

Codes are shared across the binary and match `internal/exitcodes`.

## 7. Error message style, help, and discoverability

Error style:

```text
Error: <noun> "<identifier>": <reason>
```

The `Error:` line always names the offending URL, path, flag, command, or config key literally. There are no separate `Hint:` lines. When remediation is important, it is included in the reason sentence.

Help text drafts:

`gcm` short: Manage cloned git repositories

`gcm` long:

```text
Manage cloned git repositories by deriving clone paths, scanning local repository status, and showing lightweight configuration.

Examples:
  gcm clone git@github.com:nWave-ai/aihero.git
  gcm status
  gcm status --no-fetch --non-default
  gcm config show

Docs: docs/prd/v1-core-commands.md
Issues: docs/issues/
```

`gcm clone` short: Clone a repository into its derived path

`gcm clone` long:

```text
Clone a repository to the path derived from the clone root and the repository URL hierarchy.

Examples:
  gcm clone git@github.com:nWave-ai/aihero.git
  gcm clone https://github.com/nWave-ai/aihero.git

Docs: docs/prd/v1-core-commands.md
Issues: docs/issues/
```

`gcm status` short: Show repository status under the clone root

`gcm status` long:

```text
Scan every git repository under the clone root and show branch, behind, dirty, and badge status.

Examples:
  gcm status
  gcm status --no-fetch
  gcm status --no-fetch --non-default

Docs: docs/prd/v1-core-commands.md
Issues: docs/issues/
```

`gcm config set clone-root` short: Set the clone root path

`gcm config set clone-root` long:

```text
Write the clone root override used by clone and status commands.

Examples:
  gcm config set clone-root ~/src/work
  gcm config show

Docs: docs/prd/v1-core-commands.md
Issues: docs/issues/
```

`gcm config show` short: Show the effective configuration

`gcm config show` long:

```text
Show effective configuration values, including built-in defaults when no config file sets a value.

Examples:
  gcm config show
  GCM_CONFIG=/tmp/gcm.yaml gcm config show

Docs: docs/prd/v1-core-commands.md
Issues: docs/issues/
```

Discoverability:

- No-arg behavior: `gcm` prints concise top-level help and exits 0; this is friendlier for a small tool than treating no args as an error.
- Group-only behavior: `gcm config` and `gcm config set` print their group help and exit 0.
- Missing-input behavior on non-interactive stdin: print the canonical error and exit 2; do not dump full help.
- `help` subcommand: yes, Cobra's `gcm help` and `gcm help <command>` are part of the v1 contract.
- Completion: `gcm completion bash|zsh|fish|powershell` is part of the v1 contract through Cobra.
- Man pages: no man pages in v1.
- Tree listing: `clone`, `status`, `config`, `help`, and `completion` appear in root help; config leaves appear in group-level help.

## 8. Confirmation, dry-run, idempotence

| Subcommand | Writes or destroys state? | Idempotence | Confirmation | Dry run | Force |
|---|---:|---|---|---|---|
| `gcm clone` | yes | Yes-with-caveats: rerun skips existing git repo; blocked non-git path fails | none; clone is non-destructive and blocked destinations fail | no | no |
| `gcm status` | no, except git fetch updates remote refs | Yes-with-caveats: remote refs may change because fetch is default | none | no; use `--no-fetch` to avoid network/update side effects | no |
| `gcm config set clone-root` | yes | Yes: same path produces same config value | none; explicit set command is the confirmation | no | no |
| `gcm config show` | no | Yes | none | n/a | no |

No v1 command supports `--dry-run`.

## 9. Stability promises

| Element | Stability |
|---|---|
| Subcommand names + tree position | Contract |
| Long flag names | Contract |
| Short flag names | Contract |
| Exit codes | Contract |
| Error message stderr substrings | Convenience |
| Plain default output text | Convenience |
| stderr diagnostic text | Convenience |

## 10. Color and TTY behavior

| Token | ANSI sequence | Screens class / alias |
|---|---|---|
| `Error:` | red `\e[31m...\e[0m` | `error-label` |
| `[behind]` | yellow `\e[33m...\e[0m` | `badge-behind` |
| `[!<default_branch>]` | blue `\e[34m...\e[0m` | `badge-non-default` |
| `[fetch-failed]` | red `\e[31m...\e[0m` | `badge-fetch-failed` |
| `[no-remote]` | magenta `\e[35m...\e[0m` | `badge-no-remote` |

Color is emitted only when the relevant stream is a TTY and `NO_COLOR` is unset. stdout and stderr are checked independently. v1 does not add `--no-color` or `--color`; automatic TTY detection plus `NO_COLOR` is enough. Cursor movement, screen clears, line redraws, spinners, and progress bars are out of scope.

## 11. `output-spec.md` compliance

| Spec rule | Status / Exception |
|---|---|
| stdout reserved for results; no logs/progress/diagnostics on stdout | OK |
| stderr used for diagnostics, warnings, and errors | OK |
| commands do not interleave diagnostics into stdout | OK |
| only plain text output is supported | OK |
| `--format`, `--output`, and structured-output flags are not part of v1 | OK |
| errors use one-part `Error:` form | OK |
| no separate `Hint:` lines are emitted | OK |
| exit codes are 0 success, 1 general/partial failure, 2 usage/config | OK |
| color only when stream is a TTY and `NO_COLOR` is unset | OK |
| v1 does not provide `--color` or `--no-color` | OK |
| v1 does not support dry-run flags | OK |

## 12. Open questions for downstream work

- `gcm status` tips mention `gcm pull`, which is post-v1. The PRD requires the tip; downstream work should decide whether this is acceptable as a forward pointer.
- A binary-specific `docs/specs/gcm-cmd-output-spec.md` is not needed for v1, but may become useful once more commands add specialized behavior.
