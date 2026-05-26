# 0015 — `gcm open`: interactive repository selection with cd-only wrapper

Status: ready-for-agent

## What to build

A new `gcm open` command that lets the user pick a repository under the **Clone Root** via fzf and cd into it. This slice ships the full end-to-end path *without* opener-launch — that lands in 0016. The `open` branch of the shell-init wrapper performs only the cd.

**Go-side `gcm open`**:

- Enumerates repositories under the Clone Root by reusing `internal/repositorywalker`. If the walker today is coupled to status collection, add a paths-only mode that returns absolute repository paths and nothing else.
- Spawns `fzf` as a subprocess. The candidate stream is one absolute repository path per line, sorted, deduplicated. fzf's UI runs on the controlling TTY; its stdout (the selection) is captured by gcm.
- The `--preview` argument is a shell snippet ported literally from the user's `proj()` function: try README in six casings/extensions, prefer `bat` if present, fall back to `head -n 40`, print `No README found in $dir` if none. Preview window: `right:60%`.
- On selection, prints the selected absolute path to stdout and exits 0. Nothing else on stdout.
- Optional positional argument: `gcm open <query>` passes `--query <query>` to fzf to pre-fill the filter. Zero args → no `--query`.
- If `fzf` is not on `$PATH`: exit nonzero with an error directing the user to the fzf install page. No fzf detection at shell-init time.
- If the walker returns zero repositories: exit 1 with an actionable error stating the clone root path and pointing at `gcm clone`. Distinguish from "clone root does not exist" (separate, equally actionable error).
- If the user aborts fzf (Esc / Ctrl-C): exit 0 with empty stdout. The wrapper's `[ -n "$dir" ]` guard skips the cd.
- When run *without* the shell wrapper, behaviour is unchanged for scripts (path on stdout, exit 0) but a stderr hint is emitted *only when stdout is a TTY*, mirroring the `gcm shell-init` discoverability pattern from PRD story 44a. Hint points at `eval "$(gcm shell-init)"`.

**shell-init wrapper**:

- The wrapper function gains an `open` branch alongside the existing `clone` branch, in both the POSIX (bash/zsh) and fish variants.
- The `open` branch: `dest=$(command gcm open "$@")`; if exit 0 and `-n "$dest"`, `cd "$dest"`. No opener launch in this slice — that arrives in 0016.
- The `clone` branch is untouched.

## Acceptance criteria

- [ ] `gcm open` walks the configured Clone Root and presents repositories in fzf, with the proj-style README preview on the right
- [ ] Selecting a repository in fzf prints its absolute path to stdout and exits 0
- [ ] Aborting fzf (Esc/Ctrl-C) exits 0 with empty stdout
- [ ] `gcm open <query>` pre-fills the fzf filter via `--query`
- [ ] `gcm open` with `fzf` missing exits nonzero with an actionable error
- [ ] `gcm open` against an empty clone root exits 1 with an error naming the clone root path and pointing at `gcm clone`
- [ ] `gcm open` against a non-existent clone root exits 1 with a distinct error pointing at `gcm config set clone-root`
- [ ] After `eval "$(gcm shell-init)"`, `gcm open` followed by selection cd's the shell into the selected repository
- [ ] After `eval "$(gcm shell-init)"`, aborting fzf leaves the shell's cwd unchanged
- [ ] Running `gcm open` bare (no wrapper) in an interactive terminal emits a stderr install hint; piping/capturing stdout suppresses the hint
- [ ] Unit tests cover: walker paths-only mode, zero-repos error, missing-fzf error, abort path (exit 0 + empty stdout), TTY-gated install hint
- [ ] The `clone` wrapper branch and existing clone behaviour are unchanged

## Blocked by

- None - can start immediately
