# CLI Surface: Actual CLI Screens and Errors

Generated: 2026-05-26T14:16Z
Last regenerated: 2026-05-26T14:16Z

## 0. Source artifacts

- Input: inline request, "generate documentation of actual screens and error messages"
- Binary: `gcm` (entry point: `cmd/gcm/main.go`; command tree: `internal/cmd/`)
- Conventions consulted:
  - `CONTEXT.md`
  - `README.md`
  - `docs/specs/output-spec.md`
  - `docs/specs/gcm-cmd-output-spec.md`: absent - author as follow-up only if per-command elaboration becomes necessary
  - existing Cobra tree under `internal/cmd/`
  - actual command probes, including `gcm --help`, subcommand help, temp config, temp local git remotes, and failure cases
- Hint system: optional/absent. `docs/specs/output-spec.md` explicitly says there is no hint system, so actual screens do not render `Hint:` lines.
- Note on sample values: temp paths from probes are normalized to stable example paths such as `/tmp/gcm-demo`; command shapes, stream placement, wording, and exit codes match observed behavior.

## 1. Output variants

Variants are ordered by the chronological path a new user is most likely to follow: discover the tool, inspect/set config, clone, check status, install shell integration, then encounter common failures.

| # | Variant | Type | Surface elements implied |
|---|---|---|---|
| V1 | Discover root help | discoverability | `gcm` / `gcm --help`, root help text, exit 0 |
| V2 | Inspect command help | discoverability | `gcm clone --help`, `gcm status --help`, `gcm config --help`, `gcm shell-init --help` |
| V3 | Show default configuration | happy path | `gcm config show`, default clone root annotation |
| V4 | Set clone root | state change | `gcm config set clone-root <path>`, config file result line |
| V5 | Show custom configuration | happy path | `gcm config show`, configured clone root |
| V6 | Clone repository - new clone success | happy path | clone-root creation diagnostics, clone progress on stderr, destination path on stdout |
| V7 | Clone repository - already present with matching origin | idempotent happy path | destination path on stdout, no progress stderr, exit 0 |
| V8 | Status - no repositories | happy path | empty status table state and summary |
| V9 | Status - local scan with non-default dirty repository | happy path | `--no-fetch`, aligned table row, `[!main]` badge, summary, tip |
| V10 | Status - non-default filter | happy path | `--no-fetch --non-default`, filtered table, no tip |
| V11 | Shell init - print wrapper | happy path | `gcm shell-init zsh`, shell function on stdout |
| V12 | Shell init - install success | state change | `--install`, rc-file status on stderr, exit 0 |
| V13 | Shell init - install already present | idempotent happy path | idempotent stderr status, exit 0 |
| V14 | Clone repository - missing URL | usage error | Cobra positional-arg error, exit 2 |
| V15 | Clone repository - malformed URL | error path | repository URL parse error, exit 1 |
| V16 | Clone repository - destination blocked | error path | blocked destination error, exit 1 |
| V17 | Shell init - unsupported shell | usage error | `Error: shell "<name>": unsupported shell`, exit 2 |
| V18 | Unknown command | usage error | Cobra unknown-command error, exit 2 |
| V19 | Unknown flag | usage error | Cobra unknown-flag error, exit 2 |

## 2. Primary human workflow

Most common user goal: A developer checks the tool, clones a repository into the derived path under the clone root, checks local repository status, and installs shell integration for future clone auto-cd behavior.

Shortest happy-path invocation:

```console
$ gcm clone file://localhost/tmp/gcm-demo/remote/aihero.git
```

Input acquisition:

| Input | Strategy | Notes |
|---|---|---|
| repository URL | positional (arg 1) | required; passed to git as-is |
| clone root | env-var | `GCM_CONFIG` can point to config file; otherwise effective `clone_root` defaults to `~/src` |
| destination path | n/a | derived from clone root and parsed URL hierarchy |
| shell | positional | optional for `gcm shell-init`; explicit `zsh`/`bash`/`fish` avoids relying on `$SHELL` |
| confirmation | n/a | v1 commands do not use confirmation prompts |

First success line:

```text
E| Clone root /tmp/gcm-demo/src does not exist - creating it
E| Cloning to /tmp/gcm-demo/src/localhost/tmp/gcm-demo/remote/aihero...
E| Done.
S| /tmp/gcm-demo/src/localhost/tmp/gcm-demo/remote/aihero
```

First failure line for the most likely error:

```text
E| accepts 1 arg(s), received 0
```

## 3. Command tree

Layout: Existing tree is verb-first at the root for actions (`clone`, `status`, `shell-init`) and noun-first under `config`. This document records the actual current tree; no commands are renamed or moved.

Subcommands:

| Path | Purpose | Variants covered |
|---|---|---|
| `gcm` | Show root help when invoked with no arguments | V1 |
| `gcm clone <url>` | Clone a repository into its derived path and print that path as the result | V2, V6, V7, V14, V15, V16 |
| `gcm status` | Show repository status under the clone root | V2, V8, V9, V10, V19 |
| `gcm config show` | Show effective configuration, including defaults | V2, V3, V5 |
| `gcm config set clone-root <path>` | Persist the clone root override | V2, V4 |
| `gcm shell-init [--install] [bash|zsh|fish]` | Print or install shell integration for clone auto-cd behavior | V2, V11, V12, V13, V17 |
| `gcm help [command]` | Show help for root or a subcommand | V1, V2 |
| `gcm completion bash|zsh|fish|powershell` | Generate shell completion scripts | V1 |

Aliases: none observed.

Tree listing from actual `gcm --help`:

```text
Available Commands:
  clone       Clone a repository into its derived path
  completion  Generate the autocompletion script for the specified shell
  config      Manage gcm configuration
  help        Help about any command
  shell-init  Print shell integration for changing directory after clone
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

Short flags: only Cobra's built-in `-h`. No `--format`, `--output`, `--dry-run`, `--color`, or `--no-color` flags exist in v1.

## 5. Output schema and stream contract

Output mode:

| Subcommand | Default-format result | Rationale | Structured schema ref | Side-effect-only? |
|---|---|---|---|---|
| `gcm` | root help on stdout | a small command tree is easiest to discover from the terminal in one screen | n/a - no JSON output | no |
| `gcm clone` | one bare destination path on stdout; progress on stderr | the path is capturable by shell wrappers and scripts while progress remains human-only diagnostics | n/a - no JSON output | no |
| `gcm status` | aligned table with summary and optional tip | repository state scans fastest as rows with stable columns and text badges | n/a - no JSON output | no |
| `gcm config set clone-root` | `Config saved to <path>` on stdout | confirms the write and names the config file touched | n/a - no JSON output | no |
| `gcm config show` | key/value lines with default annotation when applicable | config has few fields; labelled lines are clearer than a table | n/a - no JSON output | no |
| `gcm shell-init` | shell-function wrapper script | stdout is code intended for `eval`, so it must contain only the wrapper | n/a - no JSON output | no |
| `gcm shell-init --install` | install status on stderr; no stdout in observed behavior | current implementation treats install as a user-facing mutation with diagnostic status only | n/a - no JSON output | no; it emits stderr status |

TTY-aware behavior:

| Subcommand | TTY-aware behavior |
|---|---|
| `gcm clone` | no color; stdout path and stderr progress are unchanged when piped |
| `gcm status` | badge text is always emitted; badge color is added only when stdout is a TTY and `NO_COLOR` is unset |
| `gcm config set clone-root` | no color; output unchanged when piped |
| `gcm config show` | no color; output unchanged when piped |
| `gcm shell-init` | no color; when stdout is a TTY, stderr also prints install guidance; when stdout is not a TTY, that guidance is suppressed |
| `gcm shell-init --install` | no color; status is written to stderr |

JSON schemas: none. `docs/specs/output-spec.md` says `gcm` does not support JSON output in v1.

Output-spec compliance for actual observed behavior:

- stdout vs stderr: normal command results and help go to stdout; progress, install status, and errors go to stderr.
- bare action result: `gcm clone` emits exactly one bare destination path on stdout on success, including idempotent reruns.
- structured output: no JSON or alternate output format is part of the v1 contract.
- observed mismatch: `gcm shell-init --install` writes no stdout path even though the earlier surface design expected one bare rc path. The current implementation writes install status to stderr only.
- observed mismatch: several errors are raw Cobra/runtime strings rather than the `Error: <noun> "<identifier>": <reason>` form required by `output-spec.md`.

## 6. Exit-code map

| Code | Class | Trigger | stderr substring (canonical) |
|---|---|---|---|
| 0 | success | normal completion, help, idempotent clone skip with matching origin, `shell-init --install` rerun, and empty status scans | - |
| 1 | general failure | malformed URL, blocked destination, git clone failure, status collection failure, partial fetch failure, or rc file write failure | varies; observed examples include `parse repository URL "...": missing hostname` and `cannot clone to ...` |
| 2 | usage/config error | missing argument, unknown command, unknown flag, unsupported shell, `$SHELL` unset, config errors | varies; observed examples include `accepts 1 arg(s), received 0`, `unknown command "frobnicate" for "gcm"`, `unknown flag: --json`, and `Error: shell "tcsh": unsupported shell` |

Codes are shared across the binary and match `internal/exitcodes`.

## 7. Error message style, help, and discoverability

Specified error style:

```text
Error: <noun> "<identifier>": <reason>
```

Observed actual behavior is mixed:

| Case | Actual first stderr line | Exit |
|---|---|---:|
| missing clone URL | `accepts 1 arg(s), received 0` | 2 |
| malformed URL | `parse repository URL "not-a-url": missing hostname` | 1 |
| blocked destination | `cannot clone to /tmp/gcm-demo/src/github.com/example/blocked: destination exists but is not a git repository. Move or remove it first, then run gcm clone again` | 1 |
| unsupported shell | `Error: shell "tcsh": unsupported shell` | 2 |
| unknown command | `unknown command "frobnicate" for "gcm"` | 2 |
| unknown flag | `unknown flag: --json` | 2 |

Help text observed:

`gcm` short: Manage cloned git repositories

`gcm clone` short: Clone a repository into its derived path

`gcm status` short: Show repository status under the clone root

`gcm config` short: Manage gcm configuration

`gcm shell-init` short: Print shell integration for changing directory after clone

Discoverability:

- No-arg behavior: `gcm` prints concise top-level help and exits 0.
- Group-only behavior: `gcm config` and `gcm config set` print group help and exit 0.
- Missing-input behavior on non-interactive stdin: print the Cobra positional-arg error and exit 2.
- `help` subcommand: yes, Cobra's `gcm help` and `gcm help <command>` are part of the v1 surface.
- Support/docs path: README points to `docs/cli-surface/v1-core-commands-screens.html`, `docs/prd/v1-core-commands.md`, `docs/issues/`, and `CONTEXT.md`.
- Completion: `gcm completion bash|zsh|fish|powershell` is part of the v1 contract through Cobra.
- Man pages: no man pages observed.

## 8. Confirmation, dry-run, idempotence

| Subcommand | Writes or destroys state? | Idempotence | Confirmation | Dry run | Force |
|---|---:|---|---|---|---|
| `gcm clone` | yes | Yes-with-caveats: rerun prints the destination path and exits 0 only when existing git repo origin matches; blocked or mismatched destinations fail | none; clone is non-destructive and conflicting destinations fail | no | no |
| `gcm status` | no, except git fetch updates remote refs unless `--no-fetch` is used | Yes-with-caveats: remote refs may change because fetch is default | none | no; use `--no-fetch` to avoid network/update side effects | no |
| `gcm config set clone-root` | yes | Yes: same path produces same config value | none; explicit set command is the confirmation | no | no |
| `gcm config show` | no | Yes | none | n/a | no |
| `gcm shell-init` | no | Yes | none | n/a | no |
| `gcm shell-init --install` | yes | Yes: rerun leaves an existing install line unchanged and exits 0 | none; explicit `--install` is the confirmation | no | no |

No v1 command supports `--dry-run`, `--yes`, or `--force`.

## 9. Stability promises

| Element | Stability |
|---|---|
| Subcommand names + tree position | Contract |
| Long flag names | Contract |
| Short flag names | Contract |
| Exit codes | Contract |
| Plain stdout shape | Convenience |
| Error wording | Convenience |
| stderr diagnostic wording | Convenience |

## 10. Color and TTY behavior

| Token | ANSI sequence | Screens CSS class / alias |
|---|---|---|
| `[behind]` status badge | yellow (`\e[33m...\e[0m`) | `badge-behind` |
| `[!<default-branch>]` status badge | blue (`\e[34m...\e[0m`) | `badge-non-default` |
| `[fetch-failed]` status badge | red (`\e[31m...\e[0m`) | `badge-fetch-failed` |
| `[no-remote]` status badge | magenta (`\e[35m...\e[0m`) | `badge-repo-warning` |
| `[default-unknown]` status badge | magenta (`\e[35m...\e[0m`) | `badge-repo-warning` |
| `Error:` label where present | none from implementation; rendered red in review screens for readability only when present as literal text | `error-label` |

Color is emitted only when the relevant stream is a TTY and `NO_COLOR` is unset. stdout and stderr TTY checks are independent. v1 does not provide `--color`, `--no-color`, or equivalent color-control flags. Cursor movement, screen clears, line redraws, spinners, and progress-bar animation are out of scope and not observed.

## 11. `output-spec.md` compliance

| Spec rule | Status / Exception |
|---|---|
| stdout reserved for command results; no logs/progress/diagnostics on stdout | ✅ for probed command results; help also prints to stdout as discoverability output |
| stderr used for progress/diagnostics; not a stability contract | ✅ clone progress and shell-init install status are stderr |
| success output should be plain text; no JSON in v1 | ✅ no JSON flags or JSON output observed |
| commands must not interleave diagnostics into stdout | ✅ no probed diagnostic appeared on stdout |
| errors are written to stderr | ✅ |
| errors use one-part `Error: <noun> "<identifier>": <reason>` form | ⚠ Exception: actual behavior is mixed. Unsupported shell matches this style, but Cobra positional errors, unknown command, unknown flag, URL parse errors, and blocked destination errors do not use the specified `Error:` noun form. |
| every error names the offending identifier literally | ⚠ Exception: most observed errors name the command, flag, URL, shell, or path; `accepts 1 arg(s), received 0` does not name the missing `url` argument. |
| no separate `Hint:` lines | ✅ |
| exit codes are `0`, `1`, and `2` | ✅ |
| color only when stream is TTY and `NO_COLOR` unset | ✅ for status formatter implementation |
| v1 does not provide color-control flags | ✅ |
| v1 does not support dry-run flags | ✅ |
| no-argument root invocation prints top-level help and exits 0 | ✅ |
| group-only invocation prints group help and exits 0 | ✅ |
| missing required input prints an error to stderr and exits 2 | ✅ |
| Cobra `help` and `completion` commands are part of the v1 surface | ✅ |

## 12. Open questions for downstream work

- Should actual errors be normalized to the `output-spec.md` style, or should the spec relax to allow Cobra-native messages for usage errors?
- Should `gcm shell-init --install` print the affected rc file path on stdout as a command result, or is the current stderr-only install status the intended contract?
- Should `gcm clone` blocked-destination and URL-parse errors use `Error:` prefixes for consistency with `shell-init` usage errors?
