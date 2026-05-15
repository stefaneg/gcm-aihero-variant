# git-clone-manager

A CLI tool for a single developer to clone, organise, and monitor many git repositories across one or more hosting services.

## Language

**Repository**:
A local git repository discovered under the clone root.
_Avoid_: Repo, project

**Clone Root**:
The root directory under which all repositories are organised, mirroring the remote URL hierarchy. Defaults to `~/src`.
_Avoid_: Source root, workspace root, working directory

**Status Table**:
The primary output of `gcm status`: one row per repository showing derived path (relative to clone root), current branch, commits behind, and dirty file count, preceded by a header and followed by a summary line.
_Avoid_: Status report, status output, status view

**Derived Path**:
The local filesystem path computed from a repository URL as `${clone_root}/${hostname}/${path_prefix}/${repo_name}`, mirroring the full remote URL hierarchy.
_Avoid_: Local path, target path, clone path

**Path Prefix**:
The URL path segments between the hostname and the final repository name segment. Platform-agnostic — what each platform calls these segments (org, namespace, group, owner) is an internal detail of the hosting platform module.
_Avoid_: Org path, namespace, group, owner

**Dirty**:
A repository that has uncommitted local changes (tracked or untracked files modified since the last commit).
_Avoid_: Modified, uncommitted, changed

**Default Branch**:
The branch a repository's remote designates as its primary branch, read from `refs/remotes/origin/HEAD`. Not assumed to be `main` — varies per repository.
_Avoid_: Main branch, master branch

**Behind**:
A repository whose local default branch has remote commits not yet pulled.
_Avoid_: Stale, outdated, out of date

**Discover**:
The operation of enumerating all repositories in a remote org or group via the hosting platform API, then cloning those not already present under the clone root. Post-v1.
_Avoid_: Bulk clone, sync, fetch

**Hosting Platform**:
A hostname-keyed configuration entry that records the platform type (`github`, `gitlab`, `generic`), optional protocol override, and optional API token for one hostname. Optional and post-v1.
_Avoid_: Host, host profile, remote

## Relationships

- A **Hosting Platform** is keyed by exactly one hostname (e.g., `github.com`)
- A **Repository** URL contains a hostname that may or may not have a corresponding **Hosting Platform** entry
- All **Repositories** live under a single **Clone Root**

## Example dialogue

> **Dev:** "When I run `gcm status`, does it only show repositories that gcm cloned?"
> **Domain expert:** "No — it shows every repository found under the clone root, regardless of how it got there."

> **Dev:** "What if a repository is on a feature branch and also behind?"
> **Domain expert:** "Both badges appear on the same row. It sorts as non-default-branch first, so it rises to the top regardless of how many commits behind it is."

> **Dev:** "Do I need a hosting platform configured before I can clone?"
> **Domain expert:** "No — a hosting platform entry is optional. Without one, the URL is passed to git as-is and the derived path is computed from the hostname, path prefix, and repository name in the URL."

> **Dev:** "What's the difference between a dirty repository and one that's behind?"
> **Domain expert:** "Dirty means there are local uncommitted changes. Behind means the remote default branch has commits the local branch hasn't pulled. A repository can be both."

## Flagged ambiguities

- `--non-main` was used throughout early drafts — resolved: renamed to `--non-default` to match the **Default Branch** term and correctly handle repositories where the default branch is not `main`.
- "org path", "namespace", "group", "owner" were used interchangeably for URL path segments between hostname and repository name — resolved: **Path Prefix**. Platform-specific terms are confined to hosting platform modules.
