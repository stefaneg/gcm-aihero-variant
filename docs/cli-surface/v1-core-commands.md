# CLI Surface: v1 Core Commands

Generated: 2026-05-25T13:27Z
Last regenerated: 2026-05-26T12:47Z

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
| V1 | Clone repository - new clone success | happy path | `gcm clone <url>`, bare destination path on stdout, progress on stderr |
| V2 | Clone repository - already present with matching origin | idempotent happy path | bare destination path on stdout, no progress stderr, exit 0 |
| V3 | Clone repository - existing git repo with different origin | error path | origin mismatch error, exit 1 |
| V4 | Clone repository - destination blocked | error path | non-empty/non-git destination conflict, exit 1 |
| V5 | Clone repository - empty directory target | happy path | pre-existing empty destination accepted, bare destination path result |
| V6 | Clone repository - invalid or missing URL | error path | URL positional validation, exit 2 for missing, exit 1 for malformed |
| V7 | Status - default table | happy path | `gcm status`, fetch by default, tiered table, badges, summary, tip |
| V8 | Status - no fetch | happy path | `--no-fetch`, local-state table |
| V9 | Status - non-default filter | happy path | `--non-default`, filtered table |
| V10 | Status - partial fetch failure | partial failure | table with `[fetch-failed]`, stderr error, exit 1 |
| V11 | Status - no repositories | happy path | empty clone-root table state, summary, exit 0 |
| V12 | Status - non-default filter matches zero rows | happy path | filter-aware message, total/non-default summary, no tip |
| V13 | Config set clone-root - success | state change | `gcm config set clone-root <path>`, config path result |
| V14 | Config set clone-root - missing path | error path | required positional, exit 2 |
| V15 | Config show - default plain | happy path | effective config with default annotation |
| V16 | Shell init - auto-detected shell | happy path | `gcm shell-init`, wrapper script on stdout, install hint on stderr only when stdout is a TTY |
| V17 | Shell init - explicit shell | happy path | `gcm shell-init bash|zsh|fish`, wrapper script on stdout |
| V18 | Shell init - install success | state change | `gcm shell-init --install`, rc file chosen from shell, idempotent append |
| V19 | Shell init - install already present | idempotent happy path | `--install` rerun detects existing install line and leaves file unchanged |
| V20 | Shell init - unsupported or undetectable shell | error path | literal shell/error named, exit 2 |
| V21 | Discoverability - root and group help | discoverability | `gcm`, `gcm config`, `gcm help <command>`, completion command, shell-init listing |
| V22 | Unknown command or flag | usage error | Cobra-style usage error normalized to exit 2 |

The PRD stories map directly: clone stories 1-8, 41, and 45-47 to V1-V6; status stories 9-33, 40, 48, and 49 to V7-V12; config stories 34-37 to V13-V15; shell-init stories 42-44b to V16-V20; and error/exit stories 38-39 to V3, V4, V6, V10, V14, V20, and V22.

## 2. Primary human workflow

Most common user goal: A developer clones a repository by URL and lands in the derived working directory through the shell wrapper.

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
| shell wrapper | env-var | installed separately with `eval "$(gcm shell-init)"`; clone itself does not inspect shell state |
| confirmation | n/a | clone is idempotent only when existing origin matches; conflicting destinations fail |

First success line:

```text
S| /Users/gulli/src/github.com/nWave-ai/aihero
```

First failure line for the most likely error:

```text
E| Error: destination "/Users/gulli/src/github.com/nWave-ai/aihero": exists but is not a git repository. Move or remove it, then run gcm clone again.
```

## 3. Command tree

Layout: Existing tree is verb-first at the root for actions (`clone`, `status`, `shell-init`) and noun-first under `config`. Keep that shape; no existing commands are renamed or moved.

Subcommands:

| Path | Purpose | Variants covered |
|---|---|---|
| `gcm clone <url>` | Clone a repository into its derived path and print that path as the result | V1, V2, V3, V4, V5, V6 |
| `gcm status` | Show repository status under the clone root | V7, V8, V9, V10, V11, V12 |
| `gcm config set clone-root <path>` | Persist the clone root override | V13, V14 |
| `gcm config show` | Show effective configuration, including defaults | V15 |
| `gcm shell-init [--install] [bash|zsh|fish]` | Print or install a shell-function wrapper for `gcm clone` auto-cd behavior | V16, V17, V18, V19, V20 |
| `gcm help [command]` | Show help for root or a subcommand | V21 |
| `gcm completion bash|zsh|fish|powershell` | Generate shell completion scripts | V21 |

Aliases: none. The command names are already short and script-friendly.

Tree diff against current `gcm --help`:

```diff
 Available Commands:
   clone       Clone a repository into its derived path
   completion  Generate the autocompletion script for the specified shell
   config      Manage gcm configuration
   help        Help about any command
+  shell-init  Print or install shell integration for clone auto-cd
   status      Show repository status under the clone root
```

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

`gcm shell-init [--install] [bash|zsh|fish]`:

| Long | Short | Type | Default | Required | Env var | Mutex group | Notes |
|---|---|---|---|---|---|---|---|
| `--install` |  | bool | `false` | no |  |  | append the install line to the detected or explicit shell rc file idempotently |
| positional `shell` |  | enum | auto-detect | no | `SHELL` |  | supported values: `bash`, `zsh`, `fish`; explicit positional overrides detection |

Short flags: only Cobra's built-in `-h`. No new short flags are added in v1.

The surface deliberately excludes `--format`, `--output`, `--dry-run`, `--color`, and `--no-color` per `docs/specs/output-spec.md`.

## 5. Output schema and stream contract

Output mode:

| Subcommand | Default-format result | Rationale | Structured schema ref | Side-effect-only? |
|---|---|---|---|---|
| `gcm clone` | one bare destination path on stdout; progress on stderr | the path is captured by shell wrappers and scripts, so stdout must be machine-simple while humans still see progress on stderr | n/a - no JSON output | no |
| `gcm status` | aligned table with summary and tip | many repositories are fastest to scan as rows with stable columns and text badges | n/a - no JSON output | no |
| `gcm config set clone-root` | one-line saved-path confirmation | confirms the write and names the config file touched | n/a - no JSON output | no |
| `gcm config show` | key/value lines with default annotation | config has few fields; labelled lines are easier than a table | n/a - no JSON output | no |
| `gcm shell-init` | shell-function wrapper script | the command's result is code intended for `eval`, so stdout must contain only the wrapper | n/a - no JSON output | no |
| `gcm shell-init --install` | one bare rc file path on stdout; install status on stderr | the rc path is the affected file and can be captured by scripts; human status belongs on stderr | n/a - no JSON output | no |

TTY-aware behavior:

| Subcommand | TTY-aware behavior |
|---|---|
| `gcm clone` | no color; stdout path and stderr progress are unchanged when piped |
| `gcm status` | badge text is always emitted; badge color is added only when stdout is a TTY and `NO_COLOR` is unset |
| `gcm config set clone-root` | no color; lines are unchanged when piped |
| `gcm config show` | no color; lines are unchanged when piped |
| `gcm shell-init` | no color; when stdout is a TTY, stderr also prints a discoverability line showing `eval "$(gcm shell-init)"` and `gcm shell-init --install`; when stdout is not a TTY, that hint is suppressed |
| `gcm shell-init --install` | no color; stdout path and stderr status are unchanged when piped |

Output-spec compliance for this command set:

- stdout vs stderr: command results go to stdout; diagnostics, root-creation warnings, clone progress, and errors go to stderr.
- bare action result: `gcm clone` emits exactly one bare destination path on stdout on success, including idempotent reruns.
- shell-init eval safety: `gcm shell-init` emits only wrapper code on stdout; terminal-only install guidance goes to stderr and is suppressed when stdout is not a TTY.
- structured output: no JSON or alternate output format is part of the v1 contract.

## 6. Exit-code map

| Code | Class | Trigger | stderr substring (canonical) |
|---|---|---|---|
| 0 | success | normal completion, including idempotent clone skip with matching origin, `shell-init --install` rerun, and `no-remote`/`default-unknown` status rows | - |
| 1 | general failure | git clone failure, unreadable clone root, blocked destination, origin mismatch, malformed URL, status collection failure, partial fetch failure, or rc file write failure | starts with `Error: ...` |
| 2 | usage/config error | missing argument, unknown command, unknown flag, config read/write/parse failure, unsupported shell, rc file selection failure, or `$SHELL` unset for `gcm shell-init` auto-detect | starts with `Error: ...` |

Codes are shared across the binary and match `internal/exitcodes`.

## 7. Error message style, help, and discoverability

Error style:

```text
Error: <noun> "<identifier>": <reason>
```

The `Error:` line always names the offending URL, path, flag, command, shell name, or config key literally. There are no separate `Hint:` lines. When remediation is important, it is included in the reason sentence.

Help text drafts:

`gcm` short: Manage cloned git repositories

`gcm` long:

```text
Manage cloned git repositories by deriving clone paths, scanning local repository status, and showing lightweight configuration.

Examples:
  gcm clone git@github.com:nWave-ai/aihero.git
  gcm status
  eval "$(gcm shell-init)"
  gcm config show

Docs: docs/prd/v1-core-commands.md
Issues: docs/issues/
```

`gcm clone` short: Clone a repository into its derived path

`gcm clone` long:

```text
Clone a repository to the path derived from the clone root and the repository URL hierarchy. On success stdout is exactly the destination path; progress is written to stderr.

Examples:
  gcm clone git@github.com:nWave-ai/aihero.git
  dest=$(gcm clone https://github.com/nWave-ai/aihero.git)

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

`gcm shell-init` short: Print or install shell integration for clone auto-cd

`gcm shell-init` long:

```text
Print or install a shell-function wrapper that makes gcm clone change the calling shell to the cloned repository on success. With --install, gcm appends the install line to the shell rc file if it is not already present.

Examples:
  eval "$(gcm shell-init)"
  gcm shell-init --install
  gcm shell-init zsh
  gcm shell-init --install fish

Docs: docs/prd/v1-core-commands.md
Issues: docs/issues/
```

Discoverability:

- No-arg behavior: `gcm` prints concise top-level help and exits 0; this is friendlier for a small tool than treating no args as an error.
- Group-only behavior: `gcm config` and `gcm config set` print their group help and exit 0.
- Missing-input behavior on non-interactive stdin: print the canonical error and exit 2; do not dump full help.
- `help` subcommand: yes, Cobra's `gcm help` and `gcm help <command>` are part of the v1 contract.
- Completion: `gcm completion bash|zsh|fish|powershell` is part of the v1 contract through Cobra; PowerShell completion is separate from `gcm shell-init`.
- Man pages: no man pages in v1.
- Tree listing: `clone`, `status`, `config`, `shell-init`, `help`, and `completion` appear in root help; config leaves appear in group-level help.

## 8. Confirmation, dry-run, idempotence

| Subcommand | Writes or destroys state? | Idempotence | Confirmation | Dry run | Force |
|---|---:|---|---|---|---|
| `gcm clone` | yes | Yes-with-caveats: rerun prints the destination path and exits 0 only when existing git repo origin matches; blocked or mismatched destinations fail | none; clone is non-destructive and conflicting destinations fail | no | no |
| `gcm status` | no, except git fetch updates remote refs | Yes-with-caveats: remote refs may change because fetch is default | none | no; use `--no-fetch` to avoid network/update side effects | no |
| `gcm config set clone-root` | yes | Yes: same path produces same config value | none; explicit set command is the confirmation | no | no |
| `gcm config show` | no | Yes | none | n/a | no |
| `gcm shell-init` | no | Yes | none | n/a | no |
| `gcm shell-init --install` | yes | Yes: rerun leaves an existing install line unchanged and exits 0 | none; explicit `--install` is the confirmation | no | no |

No v1 command supports `--dry-run`.

Clone cleanup statement: on clone failure, `gcm clone` cleans up only directories it created during that invocation. A user-created empty destination directory is accepted as the clone target and is not removed on failure.

Shell-init install statement: `gcm shell-init --install` chooses `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish` from the detected or explicit shell, creates parent directories when needed, appends exactly one install line, and prints the rc file path as the stdout result.

## 9. Stability promises

| Element | Stability |
|---|---|
| Subcommand names + tree position | Contract |
| Long flag names | Contract |
| Short flag names | Contract |
| `gcm clone` bare stdout path | Contract |
| `gcm shell-init` install command `eval "$(gcm shell-init)"` | Contract |
| `gcm shell-init --install` rc-file selection | Contract |
| Wrapper script internals emitted by `gcm shell-init` | Convenience |
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
| `[default-unknown]` | cyan `\e[36m...\e[0m` | `badge-default-unknown` |

Color is emitted only when the relevant stream is a TTY and `NO_COLOR` is unset. stdout and stderr are checked independently. v1 does not add `--no-color` or `--color`; automatic TTY detection plus `NO_COLOR` is enough. Cursor movement, screen clears, line redraws, spinners, and progress bars are out of scope.

## 11. `output-spec.md` compliance

| Spec rule | Status / Exception |
|---|---|
| stdout reserved for results; no logs/progress/diagnostics on stdout | OK |
| stderr used for diagnostics, warnings, errors, and progress | OK |
| action-command result on stdout is a single bare line | OK |
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

- Exact shell wrapper internals are convenience text, but implementation should test generated bash, zsh, and fish syntax with real shells where possible.
- Status formatter tests in older issues still mention a `gcm pull` tip and `main` fallback; the PRD now supersedes those older issue details.
