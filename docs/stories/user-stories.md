<!-- markdownlint-disable MD024 -->
# User Stories — git-clone-manager

## System Constraints

The following constraints apply across all stories:

1. **Platform**: CLI tool (not TUI/GUI). Command pattern: `gcm <verb> [args] [flags]`. No interactive prompts.
2. **Config location**: `~/.config/gcm/config.yaml` (XDG base dir). Overridable via `GCM_CONFIG` env var. Never auto-created — gcm operates from defaults until the user explicitly sets an override via `gcm config set`.
3. **Clone root default**: `~/src` (global default, overridable via `gcm config set clone-root`; optional per-host override in host profile). If it does not exist on first use, gcm warns the user and creates it.
4. **Path derivation**: `${clone_root}/${hostname}/${path_prefix}/${repo_name}` — mirrors the full remote URL hierarchy (GOPATH convention), including arbitrarily deep path prefixes. This is D1 from wave-decisions.md; assumption A2, risk 11.
5. **Protocol**: Passed to git as-is from the URL. gcm never rewrites the URL scheme. Host profile protocol override is optional and post-v1 only.
6. **Secrets**: API tokens and LLM API keys must never appear in stdout, stderr, debug logs, or process arguments.
7. **On-demand only**: No daemons, cron jobs, background sync, or watchers (D3, wave-decisions.md).
8. **Pull safety**: `gcm pull` never modifies repos on non-default branches or repos with uncommitted local changes (D4, wave-decisions.md). Post-v1.
9. **AI summary gated**: `gcm status --summarise` ships only after H3/T3 trust gate passes (D5, wave-decisions.md). Post-v1.
10. **Performance target**: `gcm status` must complete in under 10 seconds for 200 repos. Achieved via a parallel worker pool sized at 2× available CPU cores.
11. **Error messages**: Every error must state: what happened, why, what to do next. No raw stack traces or error codes in user-facing output.
12. **Exit codes**: 0 = success, 1 = general error, 2 = usage/config error. Partial batch success = non-zero exit.
13. **Repo discovery**: `gcm status` discovers repos via filesystem walk of `clone_root`, finding all `.git` directories. All repos found are included regardless of whether they were cloned by gcm.
14. **Default branch**: Determined per repo via `refs/remotes/origin/HEAD`; falls back to `main` if not set.
15. **Visual markers**: Status output uses text badges always (e.g., `[behind]`, `[!main]`), with color as a progressive enhancement when stdout is a TTY and `NO_COLOR` is not set.
16. **Credentials**: All git operations defer entirely to git's credential system (SSH keys, credential helpers). gcm never handles git credentials directly.
17. **DISCOVER risk acknowledgement**: All stories are grounded in single-user validation. The builder is the primary user. Evidence is directionally valid but not corroborated across 5+ interviews. Stories are validated against this known constraint.

---

## v1 Command Surface

v1 ships exactly four commands:

| Command | Description |
|---|---|
| `gcm clone <url>` | Clone a repo to its derived path under `clone_root` |
| `gcm status [--no-fetch] [--non-default]` | Cross-repo status table for all repos under `clone_root` |
| `gcm config set clone-root <path>` | Override the default clone root |
| `gcm config show` | Display effective config (defaults annotated, overrides shown) |

All other commands (host profile management, `gcm discover`, `gcm pull`, `gcm status --summarise`) are post-v1.

---

## US-01: First-Run Host Configuration

> **Post-v1.** Host profiles are optional enhancements needed only for Bulk Clone via API (US-03) and protocol override. `gcm clone` works without any host configuration. This story ships when US-03 is prioritised.

### Elevator Pitch
- **Before**: A developer installs gcm but cannot use `gcm clone` because there is no host profile. There is no guidance on what to configure or where the config file lives.
- **After**: Running `gcm config init` walks the developer through one host profile in under 2 minutes with sensible defaults shown at every prompt. The config file is created and the path confirmed.
- **Decision enabled**: The developer decides which host to configure first, which protocol to use, and whether to provide an API token now or later.

### Problem
Gudlaugur is a senior developer who just installed gcm on a new machine. He finds it frustrating to discover what needs to be configured before he can start using the tool — most CLIs fail silently on missing config or require reading docs to understand the config schema.

### Who
- Senior developer | first-run or new-machine setup | wants to configure host profiles for protocol override and API access

### Solution
A guided interactive wizard (`gcm config init`) that collects the minimum necessary config for one host, shows defaults at every step, and writes `~/.config/gcm/config.yaml`. Token is optional (required only for `gcm discover`). Completion message confirms the saved path.

### Domain Examples

#### 1: Happy path — GitHub with SSH
Gudlaugur runs `gcm config init`. He enters `github.com`, accepts the `ssh` default, enters his personal token, accepts `~/src` as clone root. Config saved. He immediately runs `gcm clone https://github.com/nWave-ai/nWave` and it works.

#### 2: Edge case — No API token provided
Gudlaugur skips the API token field. gcm creates a valid config without the token and warns: "API token not set — `gcm discover` will require one. Run `gcm config edit` to add it later." He can still use `gcm clone` immediately.

#### 3: Error — Config file write permission denied
Gudlaugur runs `gcm config init` on a system where `~/.config/` is not writable. gcm reports: "Cannot write to ~/.config/gcm/config.yaml — permission denied. Try: `mkdir -p ~/.config/gcm && chmod 700 ~/.config/gcm`."

### UAT Scenarios (BDD)

#### Scenario: Developer configures gcm for the first time
Given gcm has no configuration file at ~/.config/gcm/config.yaml  
When Gudlaugur runs `gcm config init` and provides host=github.com, protocol=ssh, token=ghp_abc123, root=~/src  
Then gcm writes a configuration file at ~/.config/gcm/config.yaml  
And the output confirms: "Config saved to ~/.config/gcm/config.yaml"  
And Gudlaugur can immediately run `gcm clone` for github.com repos

#### Scenario: Config init completes without an API token
Given gcm has no configuration file  
When Gudlaugur runs `gcm config init` and leaves the API token blank  
Then gcm creates a valid configuration without the token  
And warns that `gcm discover` will require a token  
And does not prevent any other gcm command from running

#### Scenario: Config shows defaults at every prompt
Given gcm has no configuration file  
When `gcm config init` prompts for clone root  
Then the prompt shows the default value `~/src` before the cursor  
And pressing Enter without input accepts the default

#### Scenario: Config init reports actionable error when config directory is unwritable
Given ~/.config/ is not writable by the current user  
When Gudlaugur runs `gcm config init`  
Then gcm reports: what happened (cannot write), why (permission denied), and a specific remediation command  
And exits with status 2

### Acceptance Criteria
- [ ] Config wizard prompts for: hostname, protocol (ssh/https), optional API token, clone root
- [ ] Every prompt shows a default value; pressing Enter accepts it
- [ ] The clone root prompt shows `~/src` as the default; pressing Enter records `~/src` in the config (DD8)
- [ ] Config is written to `~/.config/gcm/config.yaml` (or `$GCM_CONFIG` if set)
- [ ] Completion message includes the full path where config was saved
- [ ] Missing API token produces a warning (not an error) and allows wizard to complete
- [ ] Write failure produces an error with: what happened, why, remediation command; exits with code 2
- [ ] Running `gcm config init` when config already exists prompts before overwriting

### Outcome KPIs
- **Who**: Developer setting up host profile overrides for protocol and API access
- **Does what**: Completes host configuration without consulting documentation
- **By how much**: In under 2 minutes
- **Measured by**: Usability test timing (T1 in solution-testing.md)
- **Baseline**: No baseline — greenfield tool

### Technical Notes
- Config format: YAML (human-readable for manual editing)
- `GCM_CONFIG` env var overrides default path
- Token stored as plain text in config file; file should be created with mode 0600
- Do not accept tokens via CLI flags (TUI patterns: secrets never via flags)

---

## US-02: Clone a Single Repo to the Correct Path

### Elevator Pitch
- **Before**: Gudlaugur wants to clone a repo. He manually computes the target path and runs `git clone https://github.com/nWave-ai/nWave ~/src/github.com/nWave-ai/nWave`. One manual decision per operation.
- **After**: He runs `gcm clone https://github.com/nWave-ai/nWave`. The path is derived automatically from the URL and the output confirms the full local path.
- **Decision enabled**: Gudlaugur decides whether to clone this particular repo — not where it goes.

### Problem
Gudlaugur is a senior developer who keeps repos under `~/src/<host>/<org>/<repo>`. He finds it tedious to manually compute the correct local path every time he clones a new repo. If he makes a mistake, the repo lands in the wrong place and becomes invisible to his other tools.

### Who
- Senior developer | cloning a specific repo they've been given a URL for | wants zero path decisions

### Solution
`gcm clone <url>` derives the local path from the URL (`${clone_root}/${hostname}/${path_prefix}/${repo_name}`), passes the URL to git as-is, and clones. Output confirms the derived path before cloning starts. If the destination already exists as a git repo, skips gracefully. If `clone_root` does not exist, warns and creates it.

### Domain Examples

#### 1: Happy path — single repo cloned to derived path
Gudlaugur runs `gcm clone https://github.com/nWave-ai/nWave`. Repo clones to `~/src/github.com/nWave-ai/nWave` using https (as in the URL). Output: "Cloning to ~/src/github.com/nWave-ai/nWave..." then "Done."

#### 2: Edge case — repo already cloned
Gudlaugur accidentally re-runs `gcm clone https://github.com/nWave-ai/nWave`. The path already exists and is a valid git repo. gcm outputs "Already cloned at ~/src/github.com/nWave-ai/nWave" and exits 0. Nothing is overwritten.

#### 3: Error — destination exists but is not a git repo
Gudlaugur runs `gcm clone https://github.com/nWave-ai/nWave` but `~/src/github.com/nWave-ai/nWave` already exists as a non-git directory. gcm reports: "Cannot clone to ~/src/github.com/nWave-ai/nWave — directory exists but is not a git repository. Move or remove it first." Exits 1.

### UAT Scenarios (BDD)

#### Scenario: Repo clones to the correct derived path
Given Gudlaugur has clone root ~/src  
When he runs `gcm clone https://github.com/nWave-ai/nWave`  
Then the repo is cloned to ~/src/github.com/nWave-ai/nWave  
And the output confirms the full path before and after cloning  
And the URL is passed to git as-is without scheme rewriting

#### Scenario: Clone creates clone_root with a warning when it does not exist
Given ~/src does not exist  
When Gudlaugur runs `gcm clone https://github.com/nWave-ai/nWave`  
Then gcm warns: "Clone root ~/src does not exist — creating it"  
And creates the directory  
And proceeds with the clone

#### Scenario: Clone skips silently when destination already exists
Given ~/src/github.com/nWave-ai/nWave already exists and is a valid git repository  
When Gudlaugur runs `gcm clone https://github.com/nWave-ai/nWave`  
Then gcm skips the clone and reports "Already cloned at ~/src/github.com/nWave-ai/nWave"  
And exits with status 0  
And does not modify the existing repo

#### Scenario: Clone gives an actionable error when destination exists as a non-git directory
Given ~/src/github.com/nWave-ai/nWave exists but is not a git repository  
When Gudlaugur runs `gcm clone https://github.com/nWave-ai/nWave`  
Then gcm reports the path, explains it is not a git repository, and says to move or remove it  
And exits with status 1

#### Scenario: Cloned repo is visible in gcm status immediately after
When Gudlaugur runs `gcm clone https://github.com/nWave-ai/nWave`  
And then runs `gcm status`  
Then github.com/nWave-ai/nWave appears as a row in the status table

### Acceptance Criteria
- [ ] Path derived as `${clone_root}/${hostname}/${path_prefix}/${repo_name}`, preserving the full URL hierarchy
- [ ] URL passed to git as-is — no scheme rewriting
- [ ] Derived path displayed before cloning begins (no silent operation)
- [ ] `clone_root` does not exist → warn and create, then proceed
- [ ] Destination exists and is a git repo → skip with message, exit 0
- [ ] Destination exists and is NOT a git repo → error with explanation, exit 1
- [ ] Cloned repo appears in `gcm status` output immediately after clone

### Outcome KPIs
- **Who**: Developer cloning repos regularly
- **Does what**: Completes clone to the correct derived path without computing the path manually
- **By how much**: >80% of users complete clone to correct path in first attempt without docs (H1 target)
- **Measured by**: T1 usability task in solution-testing.md
- **Baseline**: Manual path computation — 100% of clones require manual decision

### Technical Notes
- Path derivation algorithm must be documented and stable — changes break `gcm status` for existing repos (ICP-2 in shared-artifacts-registry.md)
- Assumption A2 (URL-hierarchy model) rated risk 11 — validate in T1 usability test before treating as settled
- No dependency on US-01 — host profile is not required for clone

---

## US-03: Bulk Clone a GitHub or GitLab Org via API

> **Post-v1.** Requires host profile with API token (US-01). Ships after v1 is validated.

### Elevator Pitch
- **Before**: To clone all repos in an org, Gudlaugur writes a bespoke API script, handles auth, maps repo names to paths, and runs it once. The script does not skip already-present repos and is brittle.
- **After**: `gcm discover https://github.com/nWave-ai` fetches all repos via API, skips already-present ones, and clones the rest with progress feedback. Summary confirms exactly what happened.
- **Decision enabled**: Gudlaugur decides which org to discover. gcm handles the enumeration, de-duplication, and path placement.

### Problem
Gudlaugur is a senior developer who, when joining a new team or returning after a gap, needs to clone all repos in an org. Writing a fresh API script for each situation wastes time. His old scripts were brittle and did not track which repos were already cloned.

### Who
- Senior developer | onboarding to a new org or refreshing a local workspace | has API token for the host

### Solution
`gcm discover https://<host>/<org>` uses the configured API token to list all repos in the org/group, checks which are already cloned under `${clone_root}`, and clones the rest. Shows progress [N/M]. Reports summary: X cloned, Y skipped, Z errors. Individual failures do not abort the batch.

### Domain Examples

#### 1: Happy path — GitHub org with 12 repos, 3 already cloned
Gudlaugur runs `gcm discover https://github.com/nWave-ai`. API returns 12 repos. 3 already exist under `~/src/github.com/nWave-ai/`. 9 are cloned. Summary: "9 cloned, 3 skipped (already present), 0 errors."

#### 2: Edge case — one repo has a permissions error
During a 9-repo batch, one private repo returns 403. gcm logs: "[ERROR] nWave-ai/private-repo — forbidden (check your token permissions)". The other 8 clone successfully. Summary: "8 cloned, 3 skipped, 1 error."

#### 3: Error — API token missing
Gudlaugur runs `gcm discover https://github.com/nWave-ai` but has no API token configured for `github.com`. gcm outputs: "API token required for `gcm discover`. Run: `gcm config edit` to add a token for github.com." Exits 2.

### UAT Scenarios (BDD)

#### Scenario: Developer bulk-clones all repos in a GitHub org
Given Gudlaugur has a github.com profile with a valid API token  
And the nWave-ai org has 12 repos on GitHub, 3 of which are already cloned under ~/src  
When he runs `gcm discover https://github.com/nWave-ai`  
Then gcm clones 9 repos to their derived paths under ~/src/github.com/nWave-ai/  
And skips the 3 already-present repos  
And reports "9 cloned, 3 skipped (already present), 0 errors"

#### Scenario: Discover shows progress during a large batch
Given the nWave-ai org has 50 repos to clone  
When Gudlaugur runs `gcm discover https://github.com/nWave-ai`  
Then progress is shown as [N/50] before the first clone completes  
And the count updates for each completed clone

#### Scenario: Discover degrades gracefully when one clone fails
Given one repo in the batch has a permissions error  
When Gudlaugur runs `gcm discover https://github.com/nWave-ai`  
Then the 8 accessible repos are cloned successfully  
And the failed repo is listed with its error reason  
And the summary reflects the partial success  
And gcm exits with a non-zero status code

#### Scenario: Discover gives an actionable error when API token is missing
Given Gudlaugur has a github.com profile with no API token  
When he runs `gcm discover https://github.com/nWave-ai`  
Then gcm reports that an API token is required  
And tells him how to add one with `gcm config edit`  
And exits with status 2

### Acceptance Criteria
- [ ] Lists repos via host API (GitHub: `/orgs/{org}/repos`, GitLab: `/groups/{group}/projects`)
- [ ] Uses same path derivation algorithm as `gcm clone` (ICP-2 consistency)
- [ ] Checks existing repos under `${clone_root}` before cloning (no duplicates)
- [ ] Progress shown as `[N/M]` before first clone starts
- [ ] Individual failure does not abort the batch — logs error, continues
- [ ] Summary line always shown: "X cloned, Y skipped, Z errors"
- [ ] Missing API token → actionable error, exit 2
- [ ] API rate limit → pause with countdown, auto-resume (no user action required)
- [ ] All successfully cloned repos appear in subsequent `gcm status` output

### Outcome KPIs
- **Who**: Developer onboarding to a new org or refreshing workspace
- **Does what**: Completes bulk clone of all org repos without writing a script
- **By how much**: Entire org cloned in one command (vs. N manual clone commands or bespoke script)
- **Measured by**: Daily use count by primary user; subjective "replaced my scripts" signal
- **Baseline**: Bespoke API script per environment — ~30-60 min to write and debug

### Technical Notes
- Dependency: US-01 (host profile + API token) and US-02 (path derivation algorithm)
- API token must not appear in any output or process argument list
- Parallel clone strongly recommended for large orgs (same constraint as `gcm status` performance)
- GitLab support: groups API, not orgs API — config must distinguish host type (github/gitlab/generic)

---

## US-04: Cross-Repo Status at a Glance

### Elevator Pitch
- **Before**: Gudlaugur wants to know which of his 47 repos are behind or on non-default branches. He either opens each repo in his IDE, runs `git status` and `git fetch` per repo manually, or gives up and opens GitHub/GitLab in the browser.
- **After**: `gcm status` scans all repos in under 10 seconds and shows a table with branch, commits-behind, and uncommitted-file count for every repo. He sees the whole picture at once.
- **Decision enabled**: Gudlaugur decides which repos to investigate (behind ones), which to leave alone (clean on default branch), and which need manual attention (non-default branch with open work).

### Problem
Gudlaugur is a senior developer who manages 47-300 local repos. He finds it impossible to get a cross-repo health picture without either a manual loop through each repo or falling back to the remote UI. Task abandonment is the outcome — he confirmed this in the interview ("I gave up and used GitLab/GitHub search, which was not nice").

### Who
- Senior developer | returning from a break or starting a cross-repo investigation | wants instant visibility across all local clones

### Solution
`gcm status` performs a parallel git fetch + branch check + rev-list across all repos under `${clone_root}` (discovered via filesystem walk) and displays a table sorted by priority: non-default-branch repos first, then by commits-behind descending, then alphabetical. Paths are shown relative to `clone_root`, with the root displayed once in the header. Includes a summary line and tips for next actions. `--no-fetch` skips the network round-trip and reports from local state. `--non-default` flag filters to off-default-branch repos only.

### Domain Examples

#### 1: Happy path — 47 repos, mix of states
Gudlaugur runs `gcm status`. 3.2 seconds later, the output shows "Repos under ~/src:" followed by a 47-row table. 38 repos are on their default branch and current. 7 are behind. 5 are on non-default branches. He immediately knows what needs attention.

#### 2: Edge case — non-default-branch filter
Gudlaugur only wants to see which repos have active feature branches. He runs `gcm status --non-default`. Only the 5 non-default-branch repos appear. Summary: "5 repos on non-default branches."

#### 3: Error — one host offline
Gudlaugur runs `gcm status` when mycompany.com is offline. His 30 github.com repos show correctly. His 17 mycompany.com repos are marked `[fetch-failed]`. The summary explains partial results. He can still act on the github.com rows.

### UAT Scenarios (BDD)

#### Scenario: Status shows every local repo with branch and behind count
Given Gudlaugur has 47 repos cloned under ~/src  
When he runs `gcm status`  
Then the output header shows "Repos under ~/src:"  
And the table shows one row per repo with path relative to ~/src  
And each row includes: relative repo path, current branch, commits behind remote, uncommitted file count  
And repos behind remote show a `[behind]` badge  
And repos on non-default branches show a `[!{default_branch}]` badge using that repository's actual default branch name  
And the summary line shows total counts

#### Scenario: Status sorts non-default-branch repos first, then by commits-behind
Given 5 repos are on non-default branches and 7 repos are behind on their default branch  
When Gudlaugur runs `gcm status`  
Then the 5 non-default-branch repos appear at the top of the table  
And the 7 behind repos follow, sorted by commits-behind descending  
And repos with equal commits-behind are sorted alphabetically

#### Scenario: Status --non-default shows only repos on non-default branches
Given 5 of 47 repos are on non-default branches  
When Gudlaugur runs `gcm status --non-default`  
Then only the 5 non-default-branch repos appear  
And the summary reads "5 repos on non-default branches"

#### Scenario: Status --no-fetch reports from local state without network calls
Given Gudlaugur runs `gcm status --no-fetch`  
Then no git fetch operations are performed  
And commits-behind reflects the last known remote state  
And the command completes faster than with a fetch

#### Scenario: Status includes actionable next-step tips
Given `gcm status` output shows behind and non-default-branch repos  
Then the output includes tips pointing to `gcm pull` and `gcm status --non-default`  
And the tips reference exact command syntax

#### Scenario: Status degrades gracefully when a host is unreachable
Given mycompany.com is offline  
When Gudlaugur runs `gcm status`  
Then github.com repos show correct status  
And mycompany.com repos appear as [fetch-failed] in the table  
And the summary explains that some repos could not be reached  
And gcm exits with a non-zero status code

#### Scenario: Status marks repos with no remote configured
Given one repo has its remote removed  
When Gudlaugur runs `gcm status`  
Then that repo appears in the table marked "no-remote"  
And it is not counted in behind or current totals

#### Scenario: Status shows color badges when stdout is a TTY
Given stdout is a TTY and NO_COLOR is not set  
When Gudlaugur runs `gcm status`  
Then `[behind]` and `[!main]` badges are rendered with color  
And the text badges are always present regardless of color support

@property
#### Scenario: Status completes within time budget at scale
Given Gudlaugur has 200 repos under ~/src  
When he runs `gcm status`  
Then the command completes in under 10 seconds  
And all 200 repos appear in the output

### Acceptance Criteria
- [ ] Header shows "Repos under `${clone_root}`:" before the table
- [ ] One row per repo, columns: relative path, branch, commits-behind, uncommitted-file count
- [ ] Repos behind remote marked with `[behind]` badge
- [ ] Repos on non-default branches marked with `[!{default_branch}]` badge, where `{default_branch}` is the actual default branch name for that repository (e.g., `[!main]`, `[!master]`, `[!develop]`)
- [ ] Badges always present; color applied on top when stdout is a TTY and `NO_COLOR` is not set
- [ ] Sort order: non-default-branch first, then commits-behind descending, then alphabetical
- [ ] Summary line: "N repos — M current, P behind, Q non-default-branch"
- [ ] Summary includes tips for `gcm pull` and `gcm status --non-default`
- [ ] `--non-default` flag filters table to off-default-branch repos only
- [ ] `--no-fetch` flag skips git fetch and reports from local state
- [ ] Host unreachable → partial results with `[fetch-failed]` markers, non-zero exit
- [ ] Repo with no remote → `no-remote` marker, excluded from behind/current counts
- [ ] Default branch determined via `refs/remotes/origin/HEAD`, fallback to `main`
- [ ] Completes in <10 seconds for 200 repos (parallel worker pool at 2× CPU cores)
- [ ] All repos under `clone_root` included regardless of whether gcm cloned them
- [ ] Repos added by `gcm clone` appear in output immediately after clone

### Outcome KPIs
- **Who**: Developer checking local repo health
- **Does what**: Determines which repos are behind or on non-default branches without opening the remote UI
- **By how much**: Eliminates remote UI fallback for "what is going on locally" task (H2 target: >80% complete task using only gcm status output)
- **Measured by**: T2 usability task in solution-testing.md
- **Baseline**: Task abandonment — user "gave up and used GitLab/GitHub search"

### Technical Notes
- Parallel git fetch required to meet 10-second budget at 200 repos; worker pool = 2× CPU cores
- Performance spike needed before implementation (flagged in solution-testing.md risks)
- `behind` repo list must be consistent with `gcm pull` target selection (ICP-3 in shared-artifacts-registry.md)
- Dependency: none — clone root defaults to `~/src` without any config

---

## US-05: Bulk Pull for On-Default-Branch Repos

> **Post-v1.** Ships after Cross-Repo Status (US-04) is validated in use.

### Elevator Pitch
- **Before**: Gudlaugur wants to update all his behind repos. He either runs `git pull` in each repo directory manually, or writes a loop. He worries about accidentally pulling into a feature branch with uncommitted work.
- **After**: `gcm pull` updates all behind on-default-branch repos in one command and explicitly confirms in the output which repos on non-default branches were not touched. He can run it confidently.
- **Decision enabled**: Gudlaugur decides to update all behind repos. gcm decides which ones are safe to touch (on default branch, no uncommitted changes) and which to skip (non-default, dirty, diverged). The skip summary tells him exactly what was left alone.

### Problem
Gudlaugur is a senior developer who finds per-repo `git pull` tedious across many repos. More importantly, he does not trust bulk pull operations that might clobber active work on feature branches — he has been burned by this pattern before.

### Who
- Senior developer | wanting to refresh all behind on-default-branch repos | needs confidence that feature-branch work is untouched

### Solution
`gcm pull` fast-forward-pulls all repos that pass all three safety conditions: (1) on default branch, (2) no uncommitted local changes, (3) behind remote. Non-default-branch repos are explicitly reported as "skipped (non-default — not touched)". Diverged repos get "manual merge needed" with the repo name. Summary shows every category.

### Domain Examples

#### 1: Happy path — 7 behind on-default-branch, 5 non-default-branch
Gudlaugur runs `gcm pull`. 7 repos are fast-forward-pulled. 5 non-default-branch repos appear in the summary as "skipped (non-default — not touched)." He knows exactly what happened.

#### 2: Edge case — behind repo with uncommitted changes on default branch
One behind repo has uncommitted changes on the default branch. gcm skips it with: "[WARNING] ~/src/github.com/nWave-ai/agents — skipped (uncommitted changes). Commit or stash before pulling." The other behind repos pull normally.

#### 3: Error — diverged repo on default branch
One repo on the default branch has diverged from remote (cannot fast-forward). gcm skips it: "[WARNING] ~/src/github.com/nWave-ai/platform — skipped (manual merge needed). Run: `cd ~/src/github.com/nWave-ai/platform && git merge origin/main`." Explicit path in the hint means zero command-line thinking.

### UAT Scenarios (BDD)

#### Scenario: Pull updates all behind on-default-branch repos and leaves non-default-branch untouched
Given 7 repos are on their default branch, behind, and clean  
And 5 repos are on non-default branches  
When Gudlaugur runs `gcm pull`  
Then gcm fast-forward-pulls the 7 on-default-branch repos  
And the 5 non-default-branch repos are not modified  
And the summary reports "7 pulled, 5 skipped (non-default — not touched)"

#### Scenario: Pull skips a repo with uncommitted changes and explains why
Given a repo is on its default branch, behind, but has uncommitted local changes  
When Gudlaugur runs `gcm pull`  
Then gcm skips that repo  
And reports the specific repo path and the reason (uncommitted changes)  
And does not discard or overwrite any local changes

#### Scenario: Pull skips a diverged repo and suggests manual resolution
Given a repo on its default branch cannot be fast-forwarded (has diverged)  
When Gudlaugur runs `gcm pull`  
Then gcm skips that repo  
And reports "manual merge needed" with the specific repo path  
And does not attempt an automatic merge

#### Scenario: Pull shows progress during a large batch
Given 20 repos are behind and on their default branch  
When Gudlaugur runs `gcm pull`  
Then progress is shown as [N/20] as each pull completes  
And the summary appears after all pulls finish

### Acceptance Criteria
- [ ] Pulls only repos matching all three conditions: on default branch AND no uncommitted changes AND behind remote
- [ ] Non-default-branch repos → skip with message "skipped (non-default — not touched)"
- [ ] Repos with uncommitted changes → skip with warning and repo path
- [ ] Diverged repos → skip with "manual merge needed" and exact `git merge` command with full path
- [ ] Progress shown as `[N/M]` during the pull batch
- [ ] Summary always shows: X pulled, Y skipped (non-default), Z skipped (dirty), W skipped (diverged)
- [ ] Pulls are fast-forward only — no automatic merges
- [ ] Network failure mid-batch → report which succeeded and which failed, exit non-zero

### Outcome KPIs
- **Who**: Developer refreshing their local workspace
- **Does what**: Pulls all behind on-default-branch repos and gains confidence that feature branches are untouched
- **By how much**: Replaces per-repo `git pull` loop entirely for the update use case
- **Measured by**: Primary user adoption — does he use `gcm pull` daily vs. manual `git pull` in individual repos?
- **Baseline**: Per-repo `git pull` or manual loop — time proportional to behind repo count

### Technical Notes
- Safety rule for non-default-branch repos is D4 from wave-decisions.md — must not be weakened without evidence from new user interviews
- Dependency: US-04 (behind repo list) — pull should either reuse status output or re-fetch consistently (ICP-3)
- `git pull --ff-only` is the correct underlying operation; never `git merge` or `git pull` (which may merge)

---

## US-06: AI-Generated Change Summary for Behind Repos

### Elevator Pitch
- **Before**: Gudlaugur sees 7 behind repos in `gcm status`. To understand what changed in each, he opens git log in each repo — 7 separate operations, each requiring mental parsing of commit messages.
- **After**: `gcm status --summarise` shows a 2-3 sentence plain-English paragraph per behind repo describing what changed, so he can decide which repos deserve his attention before opening any of them.
- **Decision enabled**: Gudlaugur decides which repos to open and investigate further, guided by the AI summary rather than raw commit messages.

### Problem
Gudlaugur is a senior developer who, after checking status, still has to open git log in each behind repo to understand what changed. With 7+ behind repos, this is cognitive overhead that delays starting actual work. He mentioned unprompted: "An AI-generated summary could actually be very useful here."

### Who
- Senior developer | reviewing multiple behind repos | wants to understand changes without reading raw commit logs

### Solution
`gcm status --summarise` (opt-in flag) sends commit log and changed file names for each behind repo to an LLM and displays a human-readable paragraph per repo. Always includes a disclaimer to verify critical changes in git log. Degrades gracefully to raw commit summaries if LLM is unavailable.

**This story is separately gated**: Do not ship until H3/T3 trust validation passes (>70% trust threshold).

### Domain Examples

#### 1: Happy path — 7 behind repos, LLM available
Gudlaugur runs `gcm status --summarise`. For `nWave-ai/agents` (14 commits behind), the summary reads: "The agent framework added parallel task execution and introduced a retry policy for failed tool calls. Breaking change: tool interface now requires a context argument." He immediately knows this repo needs close review.

#### 2: Edge case — LLM API is down
Gudlaugur runs `gcm status --summarise` but the LLM endpoint is unreachable. gcm shows raw commit summaries for each repo instead, with a note: "AI summary unavailable (LLM API unreachable). Showing raw commit list." He still gets useful output.

#### 3: Error — no LLM API key configured
Gudlaugur runs `gcm status --summarise` without having set up an LLM API key. gcm outputs: "AI summaries require an LLM API key. Run: `gcm config edit` and add an `ai.api_key` entry." Exits 2.

### UAT Scenarios (BDD)

#### Scenario: AI summary provides readable narrative for each behind repo
Given Gudlaugur has 7 behind repos and a valid LLM API key in his config  
When he runs `gcm status --summarise`  
Then each behind repo shows a human-readable paragraph describing recent changes  
And each summary is labelled with the repo name and commit count behind  
And the output includes a disclaimer: "AI summaries are generated from commit messages and changed file names. Always verify critical changes in the git log."

#### Scenario: AI summary is never generated without the --summarise flag
Given Gudlaugur has a valid LLM API key configured  
When he runs `gcm status` without the --summarise flag  
Then no LLM API calls are made  
And no AI summaries appear in the output  
And the command behaves identically to pre-AI-feature status

#### Scenario: AI summary degrades to raw commits when LLM is unavailable
Given the LLM API endpoint is unreachable  
When Gudlaugur runs `gcm status --summarise`  
Then gcm shows raw commit message summaries instead of AI narratives  
And explains that AI summary is temporarily unavailable  
And does not fail with an error exit code  
And the disclaimer is adapted to note degraded mode

#### Scenario: Missing LLM API key produces an actionable error
Given no LLM API key is configured  
When Gudlaugur runs `gcm status --summarise`  
Then gcm reports that an LLM API key is required  
And provides the exact config key to set (`ai.api_key`)  
And exits with status 2

### Acceptance Criteria
- [ ] `--summarise` flag is required; feature never activates without explicit opt-in
- [ ] Per-behind-repo paragraph shown above or below the status row
- [ ] Disclaimer always present: "Always verify critical changes in the git log"
- [ ] LLM API unavailable → degrade to raw commit list + explanation, exit 0 (not an error)
- [ ] Missing LLM API key → actionable error with config key and command, exit 2
- [ ] LLM API key never appears in stdout, stderr, debug logs, or process argument list
- [ ] Summary only for behind repos — current repos do not generate LLM calls
- [ ] `gcm status` (without `--summarise`) makes no LLM calls (verified by test with network isolation)

### Outcome KPIs
- **Who**: Developer reviewing multiple behind repos
- **Does what**: Decides which repos to investigate based on AI summary, before opening git log
- **By how much**: >70% act on AI summary before checking raw git log (H3 target from solution-testing.md)
- **Measured by**: T3 Wizard-of-Oz experiment (solution-testing.md)
- **Baseline**: Assumption A4 risk score 13 — currently not validated; baseline is "everyone reads git log first"

### Technical Notes
- Gate: H3/T3 must pass before shipping (D5, wave-decisions.md)
- LLM API key stored in `~/.config/gcm/config.yaml` → `ai.api_key` (or `GCM_LLM_API_KEY` env var)
- Input to LLM: commit messages (last N commits since local HEAD) + changed file list — no code content sent
- LLM cost per repo at 200-repo scale is an unspiked risk — spike required before R3 build starts
- Dependency: US-04 (behind repo identification)

---

## Changed Assumptions

### Clarification applied: 2026-05-15

**What was clarified**: Multiple design decisions were resolved through structured grilling of the user stories. The original document assumed host profiles were a prerequisite for `gcm clone`, that a setup wizard was necessary for first use, and used the term "stale" inconsistently. Several key implementation parameters were unspecified.

**What was ambiguous**:
- Whether `gcm clone` required a host profile (protocol, config) before it could run
- Whether a setup wizard was needed to bootstrap the tool
- How repos were discovered for `gcm status` (filesystem walk vs. registry)
- Sort order for the status table
- How "stale" was defined and what threshold applied
- Whether `gcm status` fetched by default or required an explicit flag
- How the default branch was determined per repo
- What "changed files" meant in the status table
- How credentials were handled for private repo fetches
- The complete v1 command surface

**Changes made**:
1. `gcm clone` no longer requires a host profile. The URL is passed to git as-is; protocol is whatever is in the URL. Host profiles are optional, post-v1 enhancements.
2. The setup wizard (`gcm config init`) is dropped from v1. gcm operates from defaults (`~/src`) with no config file required.
3. The config file is never auto-created. It only appears when the user runs `gcm config set`.
4. "Stale" replaced with "behind" throughout — defined as any number of commits behind remote, no threshold.
5. `gcm status` fetches by default; `--no-fetch` skips the network round-trip.
6. Repo discovery uses a filesystem walk of `clone_root`. All `.git` directories are included regardless of whether gcm cloned them.
7. Sort order: non-default-branch first, then commits-behind descending, then alphabetical.
8. Paths in status output are relative to `clone_root`, with the root shown once in the header.
9. Visual markers: text badges always, color as progressive enhancement when stdout is a TTY and `NO_COLOR` is not set.
10. Worker pool for parallel operations: 2× available CPU cores.
11. Default branch per repo: read from `refs/remotes/origin/HEAD`, fallback to `main`.
12. Changed-file count in status: local uncommitted changes only (`git status --short` count).
13. Credentials: gcm defers entirely to git's credential system for all git operations.
14. v1 command surface locked to four commands: `gcm clone`, `gcm status`, `gcm config set clone-root`, `gcm config show`.
15. US-01 (First-Run Host Configuration) and US-03 (Bulk Clone) and US-05 (Bulk Pull) explicitly marked post-v1.

**What was left unchanged**: Path derivation algorithm (`${clone_root}/${hostname}/${path_prefix}/${repo_name}`), full URL hierarchy mirroring, exit code conventions, error message format (what/why/what-to-do), the AI summary gate (H3/T3), performance target (10 seconds for 200 repos), and the DISCOVER risk acknowledgement.

### Clarification applied: 2026-04-27

**What was clarified**: The default value for the clone root prompt in `gcm config init` was implicit in journey examples and TUI mockups (`~/src`) but not stated as a named, traceable decision anywhere in the specs.

**What was ambiguous**: A reader of `user-stories.md` alone could see that "every prompt shows a default" (US-01 AC) but could not determine what the clone root default actually was without cross-referencing the journey YAML. The acceptance test scenario in `config-scenarios.feature` also used the abstract phrase "set to the default value" without naming it.

**Changes made**:
1. `user-stories.md` US-01 Acceptance Criteria: added an explicit AC item naming `~/src` as the clone root default (referencing DD8).
2. `tests/acceptance/git-clone-manager/features/config-scenarios.feature`: the "Config init shows a default value at every prompt" scenario now asserts `the clone root is set to "~/src"` instead of the abstract `the default value`.
3. `wave-decisions.md`: added DD8 recording `~/src` as the explicit default clone root decision, with rationale and revisit condition.

**What was left unchanged**: All test-isolation paths (`/home/dev/repos`, `/tmp/repos`) in acceptance tests are user-supplied values, not defaults — they were correct and are untouched. The journey YAML and feature file already named `~/src` consistently and needed no changes.
