# PRD: v1 Core Commands

## Problem Statement

A senior developer managing dozens to hundreds of local git repositories across multiple hosting platforms faces three recurring friction points:

1. **Clone friction**: Computing the correct local path for a new repository is a manual, error-prone step performed every time. A wrong path makes the repository invisible to tooling that expects a consistent layout.

2. **Visibility gap**: There is no fast way to see which local repositories are behind their remotes or checked out on non-default branches. The only options are per-repository manual inspection or falling back to the hosting platform's web UI — which shows remote state, not local state.

3. **Bootstrap friction**: Most tools designed to solve these problems require configuration before they can be used, creating a setup barrier on new machines.

## Solution

`git-clone-manager` (gcm) is a zero-config CLI tool that:

- Clones repositories to a consistent derived path based on the URL hierarchy, with no configuration required
- Scans all repositories under the clone root in parallel and displays a status table showing branch, commits-behind-remote, and dirty file count for each
- Provides lightweight config commands to override the clone root default when needed

v1 ships five commands: `gcm clone`, `gcm status`, `gcm config set clone-root`, `gcm config show`, and `gcm shell-init`. The last installs a shell-function wrapper so that `gcm clone <url>` leaves the calling shell in the newly cloned working directory.

## User Stories

1. As a developer, I want to clone a repository by URL so that it lands in the correct derived path without me computing the path manually.
2. As a developer, I want the derived path to mirror the full URL hierarchy so that repositories are organised predictably regardless of path prefix depth.
3. As a developer, I want gcm to pass the URL to git as-is so that my SSH keys and credential helpers work without any additional configuration.
4. As a developer, I want gcm to warn me and create the clone root if it doesn't exist so that I don't have to create the directory manually before my first clone.
5. As a developer, I want gcm to display the derived path before cloning begins so that I can verify the destination before any files are written.
6. As a developer, I want a rerun of `gcm clone` against an already-cloned repository (matching origin URL) to print the destination path on stdout and exit 0 with no progress noise on stderr, so that script and shell-wrapper reruns behave the same as a fresh clone from the caller's point of view.
7. As a developer, I want an actionable error if the destination exists but is not a git repository so that I know what to do to resolve the conflict.
8. As a developer, I want cloned repositories to appear immediately in `gcm status` so that I can verify placement right after cloning.
9. As a developer, I want to see a status table of all repositories under my clone root so that I can assess the health of my local workspace at a glance.
10. As a developer, I want the status table header to show the clone root path so that relative paths in the table are unambiguous.
11. As a developer, I want each status table row to show the repository's current branch so that I can spot non-default-branch repositories without opening them.
12. As a developer, I want each status table row to show how many commits the repository is behind its remote so that I know which repositories need pulling.
13. As a developer, I want each status table row to show the count of dirty files so that I can see which repositories have active work in progress.
14. As a developer, I want repositories on non-default branches to appear at the top of the status table so that active work surfaces before behind main-branch repositories.
15. As a developer, I want repositories sorted by commits-behind descending within each group so that the most behind repositories are visible first.
16. As a developer, I want alphabetical ordering as the final tiebreaker in the status table so that the output is stable and predictable across runs.
17. As a developer, I want a `[behind]` badge on repositories that have remote commits not yet pulled so that I can spot problem repositories without reading the numbers.
18. As a developer, I want a `[!{default_branch}]` badge that shows the actual default branch name for each repository so that I know exactly which branch I am not on.
19. As a developer, I want badges to always appear as text so that status output is greppable and works correctly in CI and log files.
20. As a developer, I want badges to be coloured when stdout is a TTY and `NO_COLOR` is not set so that important information stands out in interactive use.
21. As a developer, I want `gcm status` to fetch from all remotes by default so that commits-behind reflects true remote state, not a cached local state.
22. As a developer, I want a `--no-fetch` flag so that I can get an instant status from local state when I do not need up-to-date remote information.
23. As a developer, I want a `--non-default` flag so that I can filter the status table to only repositories on non-default branches.
24. As a developer, I want the default branch per repository to be determined from `refs/remotes/origin/HEAD` so that gcm works correctly with repositories that use `master`, `develop`, or any other default branch name.
25. As a developer, I want gcm to treat the default branch as **unknown** when `refs/remotes/origin/HEAD` is not set so that gcm never lies about a guessed default branch, and I want a `[default-unknown]` badge on those rows so I can tell why the row is not annotated.
26. As a developer, I want `gcm status` to complete in under 10 seconds for 200 repositories so that it is practical to run routinely.
27. As a developer, I want git operations to run in parallel so that scan time scales with available CPU, not repository count.
28. As a developer, I want repositories that fail to fetch marked as `[fetch-failed]` in the status table so that I can see partial results even when a host is unreachable.
29. As a developer, I want repositories with no remote configured marked as `[no-remote]` and excluded from behind/current counts so that the summary accurately reflects fetchable repositories.
30. As a developer, I want a summary line at the bottom of the status table showing counts of current, behind, and non-default-branch repositories so that I get the overall picture without reading every row.
31. As a developer, I want the status table to include a tip pointing to `gcm status --non-default` so that I know the next useful command without consulting documentation. (A `gcm pull` tip is deferred to the release that ships `gcm pull`; advertising a non-existent command in v1 would erode trust.)
32. As a developer, I want `gcm status` to include every repository under the clone root regardless of whether gcm cloned it so that I have a complete picture of my local workspace.
33. As a developer, I want gcm to exit non-zero when any repository fails to fetch so that scripts can detect partial failures.
34. As a developer, I want to override the clone root with `gcm config set clone-root <path>` so that I can organise repositories under a non-default location.
35. As a developer, I want `gcm config show` to display the effective configuration including defaults so that I can verify what gcm is using without hunting for a config file.
36. As a developer, I want defaults annotated in `gcm config show` output so that I can distinguish configured values from fallback values at a glance.
37. As a developer, I want gcm to never auto-create a config file so that my filesystem stays clean until I explicitly configure something.
38. As a developer, I want every error message to state what happened, why, and what to do next so that I can resolve problems without consulting documentation.
39. As a developer, I want gcm to use exit code 0 for success, 1 for general errors, and 2 for config/usage errors so that scripts can handle different failure modes distinctly.
40. As a developer, I want partial batch failures to produce a non-zero exit with a partial result shown so that I know which repositories succeeded and which did not.
41. As a developer, I want `gcm clone` to print the destination path on stdout and progress messages on stderr so that a shell wrapper can capture the path with `dest=$(gcm clone <url>)` while I still see progress in my terminal.
42. As a developer, I want to install a shell-function wrapper with `eval "$(gcm shell-init)"` in my rc file so that running `gcm clone <url>` leaves my shell in the newly cloned working directory.
43. As a developer, I want `gcm shell-init` to auto-detect my shell from `$SHELL` and accept an optional positional argument that overrides detection, so that the one-line install works without thought in the common case and remains explicit when needed.
44. As a developer, I want `gcm shell-init` to support `bash`, `zsh`, and `fish` and to exit 2 with a literal error when `$SHELL` is unset or names an unsupported shell, so that I know immediately why the wrapper did not install.
44a. As a developer, I want `gcm shell-init` run bare in a terminal to print a stderr hint showing the exact line to add to my rc file (and the `--install` shortcut), so that I can discover how to install without consulting documentation. The hint is suppressed when stdout is not a TTY, so `eval "$(gcm shell-init)"` stays clean.
44b. As a developer, I want `gcm shell-init --install` to detect my shell, choose the appropriate rc file (`~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`), and append the install line idempotently, so that I do not have to edit my rc file by hand and can rerun the command safely.
45. As a developer, I want `gcm clone` to error out when the destination is an existing git repository whose `origin` URL does not match the URL I asked to clone, so that the wrapper never cds me into the wrong repository.
46. As a developer, I want `gcm clone` to accept an empty pre-existing directory as a valid destination so that I can pre-create a parent path without gcm refusing to clone.
47. As a developer, I want `gcm clone` on failure to clean up only the directories gcm itself created, so that a failed clone never leaves a broken repository on my filesystem and never deletes a directory I made.
48. As a developer, I want repositories marked `[no-remote]` or `[default-unknown]` to sort to a bottom tier of the status table (below healthy default-branch rows), so that incomplete-data rows do not crowd out actionable rows.
49. As a developer, I want `gcm status --non-default` to show a filter-aware message and a summary line with the total repository count and the non-default count when the filter matches zero repositories, so that I can distinguish "no repositories under clone root" from "filter excluded everything." The `--non-default` tip is suppressed in this state.

## Implementation Decisions

### Modules

**URL Parser**
Parses a git URL (https://, ssh://, git@) into three components: hostname, path prefix, and repository name. Pure function with no I/O — all URL format complexity is hidden behind a simple three-field output. This is the only place in the codebase that understands URL syntax.

**Path Deriver**
Given a clone root, hostname, path prefix, and repository name, computes the derived path: `${clone_root}/${hostname}/${path_prefix}/${repo_name}`. Pure function. Depends only on the URL Parser output and the configured clone root.

**Config Store**
Reads and writes `~/.config/gcm/config.yaml` (or `$GCM_CONFIG`). Exposes an effective-config view that merges file values over hard-coded defaults. Never auto-creates the config file — a write only happens when `gcm config set` is explicitly called. The only module that touches the config file on disk.

**Repository Walker**
Walks the clone root directory tree, finding all `.git` directories and returning their parent paths. Returns all repositories found, regardless of origin. No git operations — filesystem traversal only.

**Git Runner**
Executes individual git commands as subprocesses: `git clone`, `git fetch`, `git status --short`, `git rev-list`, `git symbolic-ref`, and `git remote get-url origin` (for the origin-match check in `gcm clone`). The only module that shells out to git. Takes a repository path and a command; returns structured output or a typed error. Credentials are never passed through here — they flow through git's own credential system.

**Worker Pool**
A generic parallel work runner sized at 2× the number of available CPU cores. Accepts a slice of work items and a function to apply to each; returns results in completion order with errors preserved per item. Used by the Status Command and by post-v1 Discover and Pull commands without modification.

**Status Collector**
For a single repository, uses the Git Runner to collect: current branch, default branch (via `refs/remotes/origin/HEAD`; reported as **unknown** when HEAD is unset — no `main` fallback), commits behind remote, and dirty file count. Optionally skips the fetch step when `--no-fetch` is set. Returns a typed result struct including error state (`fetch-failed`, `no-remote`, `default-unknown`). Does not sort or format.

**Status Formatter**
Takes a slice of Status Collector results, sorts them into three tiers — (1) non-default-branch rows, (2) healthy default-branch rows, (3) `[no-remote]` and `[default-unknown]` rows — then within each tier sorts by commits-behind descending and alphabetical as final tiebreaker. Formats the status table string. Applies text badges unconditionally; applies colour when stdout is a TTY and `NO_COLOR` is not set. The `[!{default_branch}]` badge uses the actual default branch name from each result — this varies per repository within a single run. Recognised badges: `[behind]`, `[!{default_branch}]`, `[fetch-failed]`, `[no-remote]`, `[default-unknown]`. When `--non-default` matches zero rows but the clone root is non-empty, the formatter emits a filter-aware message and a summary showing total and non-default counts, and suppresses the `--non-default` tip.

**Shell Init Renderer**
For a supported shell (`bash`, `zsh`, `fish`), emits the shell-function wrapper script to stdout. The wrapper intercepts `gcm clone`, captures the destination path from stdout, and `cd`s to it on success; all other subcommands pass through unchanged. Auto-detects the target shell from `$SHELL` when no positional argument is given. Exits 2 if `$SHELL` is unset or names an unsupported shell. The wrapper scripts themselves are convenience text — the install line (`eval "$(gcm shell-init)"`) is the stable contract.

### v1 command surface

| Command | Modules used |
|---|---|
| `gcm clone <url>` | URL Parser → Path Deriver → Git Runner |
| `gcm status [--no-fetch] [--non-default]` | Config Store → Repository Walker → Worker Pool → Status Collector → Status Formatter |
| `gcm config set clone-root <path>` | Config Store |
| `gcm config show` | Config Store |
| `gcm shell-init [bash\|zsh\|fish]` | Shell Init Renderer |

### Architectural decisions

- URL Parser and Path Deriver are kept strictly separate so each can be tested and reasoned about independently. The path derivation algorithm is stable and must not change without versioning — it is the key that links `gcm clone` output to `gcm status` discovery.
- Git Runner is the single point of contact with the git binary. No other module shells out to git. This boundary makes test injection straightforward.
- Worker Pool is generic and not tied to the status use case. It carries no knowledge of repositories, git, or status — just parallel execution.
- Hosting platform concepts (org, namespace, group, owner) are absent from all v1 modules. They are deferred to a post-v1 Hosting Platform module that handles API calls for Discover.
- Error messages follow a strict three-part structure (what happened / why / what to do next) enforced at the command layer, not within individual modules.
- `gcm clone` treats the destination path as its **result** and writes it on stdout as a single bare line; all progress messages ("Cloning to ...", completion notice) go to stderr. This is what makes the shell-function wrapper (`dest=$(gcm clone <url>) && cd "$dest"`) work without parsing or special modes. The bare-path stdout contract is load-bearing for shell integration and must not be decorated.
- `gcm clone` cleans up partial state on failure, but only directories it itself created during the run. A pre-existing empty directory the user made is left alone even on failure. This prevents both zombie half-clones and accidental deletion of user data.
- An empty (zero-entry) pre-existing directory at the derived path is accepted as a valid clone target, matching `git clone`'s own behaviour. Any single entry — including dotfiles — counts as non-empty and blocks the clone.
- The default branch is read from `refs/remotes/origin/HEAD` and is treated as **unknown** when HEAD is unset. gcm does not fall back to `main` because the guess is wrong often enough (master, develop, trunk) that an honest "unknown" is more useful than a confident lie.

## Testing Decisions

A good test verifies observable external behaviour — output, exit code, filesystem state — not internal implementation choices. Tests should remain valid after a refactor that preserves behaviour. Do not test private functions or assert on internal state.

**URL Parser** — unit tests covering all supported URL formats (https, ssh, git@), deep path prefixes (3+ segments), edge cases (trailing `.git` suffix, missing path prefix, bare repository names), and malformed inputs that should return errors. Pure function; no filesystem or subprocess dependencies.

**Path Deriver** — unit tests with a table of (clone_root, hostname, path_prefix, repo_name) inputs and expected derived paths. Include path prefix depths of 0, 1, 2, and 3+. Pure function.

**Status Formatter** — unit tests covering sort order (non-default-branch first, then commits-behind descending, then alphabetical), badge rendering for all states (behind, dirty, fetch-failed, no-remote), colour suppression when `NO_COLOR` is set or stdout is not a TTY, and per-repository `[!{default_branch}]` badge naming with varied default branch names across rows.

**Status Collector** — unit tests using a fake/stub Git Runner. Cover: happy path, fetch-failed, no-remote, non-default branch detection, fallback to `main` when `refs/remotes/origin/HEAD` is absent, dirty file count, `--no-fetch` skipping the fetch call.

**Config Store** — unit tests using a temp directory as config home. Cover: defaults returned when no file exists, a written value is read back correctly, `GCM_CONFIG` env var overrides the default path, file is not created on a read-only operation.

**Shell Init Renderer** — unit tests covering: each supported shell (`bash`, `zsh`, `fish`) emits a syntactically valid wrapper for that shell; auto-detection picks the right shell from `$SHELL`; explicit positional overrides detection; unset `$SHELL` with no argument exits 2; unsupported shell name (e.g. `tcsh`) exits 2 with the shell named literally in the error.

**Integration tests** — end-to-end tests that invoke the compiled binary against real local git repositories (no network). Use a temp directory as clone root. Cover the golden path for `gcm clone` (correct derived path, skip-if-exists, error-if-not-git) and `gcm status` (correct table output, `--no-fetch`, `--non-default` filter). These tests are the acceptance criteria for the commands.

## Out of Scope

- Hosting platform configuration (`gcm config init`, `gcm config add-host`) — post-v1
- Bulk discover (`gcm discover`) — post-v1, requires hosting platform API token
- Bulk pull (`gcm pull`) — post-v1
- AI-generated change summaries (`gcm status --summarise`) — gated on H3/T3 trust validation passing
- Protocol override from hosting platform config — post-v1
- GitHub / GitLab API integration — post-v1
- Interactive prompts (beyond `gcm config init` which is post-v1)
- Daemon, background sync, or file watchers
- PowerShell support in `gcm shell-init` — v1 ships bash, zsh, and fish only
- `gcm pull` status tip — reinstated when `gcm pull` ships

## Further Notes

- The project is written in Go.
- The `refs/remotes/origin/HEAD` approach to default branch detection means the default branch is always fresh after a fetch, at no extra network cost.
- The `[!{default_branch}]` badge varies per repository within a single `gcm status` run by design — it shows the actual default branch name, not a generic `[!main]` label.
- gcm never manages git credentials. This is a hard boundary that keeps the security surface minimal and avoids duplicating logic that git's credential system already handles.
- Worker pool size (2× CPU cores) and parallel fetch are implementation defaults, not user-configurable in v1. They can be exposed as config knobs in a later release if profiling reveals a need.
- The path derivation algorithm (`${clone_root}/${hostname}/${path_prefix}/${repo_name}`) must be treated as a stable contract. Any change to it breaks `gcm status` discovery for all existing clones.
