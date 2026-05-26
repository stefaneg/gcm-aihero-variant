# 0011 — `gcm shell-init` shell-function wrapper

Status: ready-for-agent

## What to build

A new top-level subcommand `gcm shell-init [bash|zsh|fish]` that prints a shell-function wrapper to stdout. Installed by the user as:

```
eval "$(gcm shell-init)"
```

in their rc file. The wrapper intercepts `gcm clone` and `cd`s the calling shell to the path that `gcm clone` emits on stdout (see issue 0010). All other subcommands pass through unchanged.

Behaviour:

- Shells supported: `bash`, `zsh`, `fish`. No others.
- When no positional argument is given, auto-detect from `$SHELL` (basename match).
- When a positional argument is given, it overrides detection.
- Unset `$SHELL` with no argument, or any unsupported shell name (e.g. `tcsh`, `sh`, `pwsh`), exits 2 with an error naming the offending value literally per `docs/specs/output-spec.md`.
- The wrapper script itself is convenience text — the install line `eval "$(gcm shell-init)"` is the stable contract. Wrapper internals can be fixed across releases without warning.
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

## Acceptance criteria

- [ ] `gcm shell-init bash` prints a bash-syntax wrapper; `gcm shell-init zsh` prints a zsh-syntax wrapper; `gcm shell-init fish` prints a fish-syntax wrapper
- [ ] `gcm shell-init` with no argument and `$SHELL=/bin/zsh` produces the zsh wrapper
- [ ] `gcm shell-init` with no argument and `$SHELL` unset exits 2 with an error
- [ ] `gcm shell-init tcsh` exits 2 with an error naming `tcsh` literally
- [ ] Integration test: in a real bash shell, `eval "$(gcm shell-init bash)"` followed by `gcm clone <local-bare-repo>` leaves cwd at the derived path; repeated invocation behaves the same
- [ ] Integration test: same flow in zsh
- [ ] Integration test: same flow in fish
- [ ] Integration test: a failing `gcm clone` (e.g., destination blocked) leaves cwd unchanged
- [ ] Integration test: `gcm status` invoked through the wrapper behaves identically to direct invocation (no cd, same output streams)

## Blocked by

- 0010 — `gcm clone` adopts stdout-as-result contract
