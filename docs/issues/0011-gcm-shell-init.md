# 0011 — `gcm shell-init` shell-function wrapper

Status: done

## What to build

A new top-level subcommand `gcm shell-init [bash|zsh|fish]` that prints a shell-function wrapper to stdout. The wrapper intercepts `gcm clone` and `cd`s the calling shell to the path that `gcm clone` emits on stdout (see issue 0010). All other subcommands pass through unchanged.

Installed by the user as one of:

- bash / zsh: `eval "$(gcm shell-init)"`
- fish: `gcm shell-init fish | source`

The user adds that line to their rc file (or uses the `--install` helper, below).

### Wrapper

Behaviour:

- Shells supported: `bash`, `zsh`, `fish`. No others.
- When no positional argument is given, auto-detect from `$SHELL` (basename match).
- When a positional argument is given, it overrides detection.
- Unset `$SHELL` with no argument, or any unsupported shell name (e.g. `tcsh`, `sh`, `pwsh`), exits 2 with an error naming the offending value literally per `docs/specs/output-spec.md`.
- The wrapper script itself is convenience text — the install line is the stable contract. Wrapper internals can be fixed across releases without warning.
- The wrapper does not modify the user's `PATH` or shadow anything other than the `gcm` invocation itself.
- On clone failure (non-zero exit from `gcm clone`), the wrapper must not cd anywhere.

Sketch of the wrapper logic (the exact form per shell will differ):

```
gcm() {
  if [ "$1" = "clone" ]; then
    local dest
    dest=$(command gcm "$@") && [ -n "$dest" ] && cd "$dest"
  else
    command gcm "$@"
  fi
}
```

`gcm completion` remains a separate subcommand; do not bundle completion output into `shell-init`.

### Interactive install hint (TTY-aware)

When `gcm shell-init` is run with stdout connected to a TTY (i.e. the user typed it bare in a terminal, not via `eval`), emit a one-paragraph hint **on stderr** showing the exact install command for the resolved shell, e.g.:

```
To install permanently, add this line to ~/.zshrc:
  eval "$(gcm shell-init)"
Or run: gcm shell-init --install
```

When stdout is not a TTY (the `eval "$(gcm shell-init)"` path), the hint is suppressed and only the wrapper passes through. This keeps the eval path clean while making the bare-invocation path self-explanatory. The hint text itself is `Convenience` per `docs/specs/output-spec.md` and may evolve.

### `--install` flag

`gcm shell-init --install` mutates the user's rc file directly. Idempotent.

- Detects shell from `$SHELL`; a positional argument overrides.
- Picks the rc file per shell:
  - `bash` → `~/.bashrc` (preferred); fall back to `~/.bash_profile` if it exists and `.bashrc` does not. Create `~/.bashrc` if neither exists.
  - `zsh` → `~/.zshrc`. Create if missing.
  - `fish` → `~/.config/fish/config.fish`. Create the directory tree if missing.
- Idempotency: scan the target file for an existing `gcm shell-init` line (any line containing the substring `gcm shell-init`). If present, exit 0 with a stderr message: `Already installed in <rc-file>.` No write.
- If absent, append the appropriate install line to the rc file with a leading blank line and a `# gcm shell-init` comment. Print a stderr confirmation: `Installed in <rc-file>. Reload your shell or run: source <rc-file>`.
- `$HOME` unset → exit 2 with a literal error naming `HOME`.
- Any write failure (permissions, disk full) → exit 1 with a literal error naming the rc file and the underlying reason.
- Stdout is empty in `--install` mode. The user wants the side effect, not the wrapper printed.

## Acceptance criteria

Wrapper rendering:

- [ ] `gcm shell-init bash` prints a bash-syntax wrapper; `gcm shell-init zsh` prints a zsh-syntax wrapper; `gcm shell-init fish` prints a fish-syntax wrapper
- [ ] `gcm shell-init` with no argument and `$SHELL=/bin/zsh` produces the zsh wrapper
- [ ] `gcm shell-init` with no argument and `$SHELL` unset exits 2 with an error
- [ ] `gcm shell-init tcsh` exits 2 with an error naming `tcsh` literally

Wrapper behaviour (integration):

- [ ] In a real bash shell, `eval "$(gcm shell-init bash)"` followed by `gcm clone <local-bare-repo>` leaves cwd at the derived path; repeated invocation behaves the same
- [ ] Same flow in zsh
- [ ] Same flow in fish (using `gcm shell-init fish | source`)
- [ ] A failing `gcm clone` leaves cwd unchanged
- [ ] `gcm status` invoked through the wrapper behaves identically to direct invocation (no cd, same output streams)

Install hint:

- [ ] `gcm shell-init` with stdout connected to a TTY prints the wrapper to stdout **and** an install-hint paragraph to stderr naming the correct rc file for the detected shell
- [ ] `gcm shell-init` with stdout redirected to a pipe or file prints only the wrapper; stderr is silent

`--install` flag:

- [ ] `gcm shell-init --install` on a fresh `$HOME` with `$SHELL=/bin/zsh` creates `~/.zshrc` and appends the install line; stderr confirms the path
- [ ] Running `--install` a second time prints `Already installed in ~/.zshrc.` on stderr, exit 0, and does not modify the file
- [ ] `--install` with `$SHELL=/bin/bash` writes to `~/.bashrc` when neither rc file exists; writes to `~/.bash_profile` when only it exists and `~/.bashrc` does not
- [ ] `--install` with `$SHELL=/usr/bin/fish` creates `~/.config/fish/` if missing and appends to `config.fish`
- [ ] `--install` with `$HOME` unset exits 2 with `HOME` named literally in the error
- [ ] `--install` against a read-only rc file exits 1 with the path and reason in the error
- [ ] `--install` produces nothing on stdout in any successful or already-installed case

## Blocked by

- 0010 — `gcm clone` adopts stdout-as-result contract
