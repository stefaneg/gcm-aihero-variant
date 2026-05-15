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

v1 ships four commands: `gcm clone`, `gcm status`, `gcm config set clone-root`, and `gcm config show`.

## User Stories

1. As a developer, I want to clone a repository by URL so that it lands in the correct derived path without me computing the path manually.
2. As a developer, I want the derived path to mirror the full URL hierarchy so that repositories are organised predictably regardless of path prefix depth.
3. As a developer, I want gcm to pass the URL to git as-is so that my SSH keys and credential helpers work without any additional configuration.
4. As a developer, I want gcm to warn me and create the clone root if it doesn't exist so that I don't have to create the directory manually before my first clone.
5. As a developer, I want gcm to display the derived path before cloning begins so that I can verify the destination before any files are written.
6. As a developer, I want gcm to skip a clone silently if the repository is already present so that I can rerun clone commands without consequences.
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
25. As a developer, I want the default branch to fall back to `main` when `refs/remotes/origin/HEAD` is not set so that gcm degrades gracefully for older repositories.
26. As a developer, I want `gcm status` to complete in under 10 seconds for 200 repositories so that it is practical to run routinely.
27. As a developer, I want git operations to run in parallel so that scan time scales with available CPU, not repository count.
28. As a developer, I want repositories that fail to fetch marked as `[fetch-failed]` in the status table so that I can see partial results even when a host is unreachable.
29. As a developer, I want repositories with no remote configured marked as `[no-remote]` and excluded from behind/current counts so that the summary accurately reflects fetchable repositories.
30. As a developer, I want a summary line at the bottom of the status table showing counts of current, behind, and non-default-branch repositories so that I get the overall picture without reading every row.
31. As a developer, I want the status table to include tips pointing to `gcm pull` and `gcm status --non-default` so that I know the next commands to run without consulting documentation.
32. As a developer, I want `gcm status` to include every repository under the clone root regardless of whether gcm cloned it so that I have a complete picture of my local workspace.
33. As a developer, I want gcm to exit non-zero when any repository fails to fetch so that scripts can detect partial failures.
34. As a developer, I want to override the clone root with `gcm config set clone-root <path>` so that I can organise repositories under a non-default location.
35. As a developer, I want `gcm config show` to display the effective configuration including defaults so that I can verify what gcm is using without hunting for a config file.
36. As a developer, I want defaults annotated in `gcm config show` output so that I can distinguish configured values from fallback values at a glance.
37. As a developer, I want gcm to never auto-create a config file so that my filesystem stays clean until I explicitly configure something.
38. As a developer, I want every error message to state what happened, why, and what to do next so that I can resolve problems without consulting documentation.
39. As a developer, I want gcm to use exit code 0 for success, 1 for general errors, and 2 for config/usage errors so that scripts can handle different failure modes distinctly.
40. As a developer, I want partial batch failures to produce a non-zero exit with a partial result shown so that I know which repositories succeeded and which did not.

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
Executes individual git commands as subprocesses: `git clone`, `git fetch`, `git status --short`, `git rev-list`, `git symbolic-ref`. The only module that shells out to git. Takes a repository path and a command; returns structured output or a typed error. Credentials are never passed through here — they flow through git's own credential system.

**Worker Pool**
A generic parallel work runner sized at 2× the number of available CPU cores. Accepts a slice of work items and a function to apply to each; returns results in completion order with errors preserved per item. Used by the Status Command and by post-v1 Discover and Pull commands without modification.

**Status Collector**
For a single repository, uses the Git Runner to collect: current branch, default branch (via `refs/remotes/origin/HEAD`, fallback `main`), commits behind remote, and dirty file count. Optionally skips the fetch step when `--no-fetch` is set. Returns a typed result struct including error state (`fetch-failed`, `no-remote`). Does not sort or format.

**Status Formatter**
Takes a slice of Status Collector results, sorts them (non-default-branch first, then commits-behind descending, then alphabetical), and formats the status table string. Applies text badges unconditionally; applies colour when stdout is a TTY and `NO_COLOR` is not set. The `[!{default_branch}]` badge uses the actual default branch name from each result — this varies per repository within a single run.

### v1 command surface

| Command | Modules used |
|---|---|
| `gcm clone <url>` | URL Parser → Path Deriver → Git Runner |
| `gcm status [--no-fetch] [--non-default]` | Config Store → Repository Walker → Worker Pool → Status Collector → Status Formatter |
| `gcm config set clone-root <path>` | Config Store |
| `gcm config show` | Config Store |

### Architectural decisions

- URL Parser and Path Deriver are kept strictly separate so each can be tested and reasoned about independently. The path derivation algorithm is stable and must not change without versioning — it is the key that links `gcm clone` output to `gcm status` discovery.
- Git Runner is the single point of contact with the git binary. No other module shells out to git. This boundary makes test injection straightforward.
- Worker Pool is generic and not tied to the status use case. It carries no knowledge of repositories, git, or status — just parallel execution.
- Hosting platform concepts (org, namespace, group, owner) are absent from all v1 modules. They are deferred to a post-v1 Hosting Platform module that handles API calls for Discover.
- Error messages follow a strict three-part structure (what happened / why / what to do next) enforced at the command layer, not within individual modules.

## Testing Decisions

A good test verifies observable external behaviour — output, exit code, filesystem state — not internal implementation choices. Tests should remain valid after a refactor that preserves behaviour. Do not test private functions or assert on internal state.

**URL Parser** — unit tests covering all supported URL formats (https, ssh, git@), deep path prefixes (3+ segments), edge cases (trailing `.git` suffix, missing path prefix, bare repository names), and malformed inputs that should return errors. Pure function; no filesystem or subprocess dependencies.

**Path Deriver** — unit tests with a table of (clone_root, hostname, path_prefix, repo_name) inputs and expected derived paths. Include path prefix depths of 0, 1, 2, and 3+. Pure function.

**Status Formatter** — unit tests covering sort order (non-default-branch first, then commits-behind descending, then alphabetical), badge rendering for all states (behind, dirty, fetch-failed, no-remote), colour suppression when `NO_COLOR` is set or stdout is not a TTY, and per-repository `[!{default_branch}]` badge naming with varied default branch names across rows.

**Status Collector** — unit tests using a fake/stub Git Runner. Cover: happy path, fetch-failed, no-remote, non-default branch detection, fallback to `main` when `refs/remotes/origin/HEAD` is absent, dirty file count, `--no-fetch` skipping the fetch call.

**Config Store** — unit tests using a temp directory as config home. Cover: defaults returned when no file exists, a written value is read back correctly, `GCM_CONFIG` env var overrides the default path, file is not created on a read-only operation.

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

## Further Notes

- The project is written in Go.
- The `refs/remotes/origin/HEAD` approach to default branch detection means the default branch is always fresh after a fetch, at no extra network cost.
- The `[!{default_branch}]` badge varies per repository within a single `gcm status` run by design — it shows the actual default branch name, not a generic `[!main]` label.
- gcm never manages git credentials. This is a hard boundary that keeps the security surface minimal and avoids duplicating logic that git's credential system already handles.
- Worker pool size (2× CPU cores) and parallel fetch are implementation defaults, not user-configurable in v1. They can be exposed as config knobs in a later release if profiling reveals a need.
- The path derivation algorithm (`${clone_root}/${hostname}/${path_prefix}/${repo_name}`) must be treated as a stable contract. Any change to it breaks `gcm status` discovery for all existing clones.
