# 0016 — `project-opener` config + opener launch baked into shell-init wrapper

Status: done

## What to build

Add the **Project Opener** concept (see CONTEXT.md and ADR 0004) so that `gcm open`, after cd'ing, also launches the user's chosen IDE/workspace command on the selected repository. The opener is configured once via `gcm config`, resolved at `shell-init` time, and baked literally into the emitted wrapper.

**Config**:

- New configstore key `project-opener` (string). Accepts a full command line verbatim — e.g. `goland`, `code --new-window`, `idea -nosplash`. No `exec.LookPath` validation at set time (mirrors how `clone-root` is accepted without filesystem validation).
- `gcm config set project-opener <cmd>` succeeds, then prints a one-line stderr hint reminding the user to re-run `eval "$(gcm shell-init)"` (or re-source their rc file) for the change to take effect in the `open` wrapper. This hint is emitted only for this key.
- `gcm config show` includes `project-opener` in its output. When unset, the default value rendered is empty (or an explicit indicator that no opener is configured) and is annotated as a default, matching the existing default-annotation pattern.
- No environment-variable fallback chain. No reading of `$EDITOR` or `$VISUAL`. The single source of truth is the config key.

**shell-init wrapper**:

- When `gcm shell-init` runs, it reads `project-opener` from the configstore *once* and bakes the result into the emitted wrapper:
  - If `project-opener` is set: the `open` branch becomes `cd "$dest" && <opener-cmd> .`, where `<opener-cmd>` is the configured value inlined verbatim.
  - If `project-opener` is unset: the `open` branch is `cd "$dest"` only, identical to 0015.
- Both the POSIX (bash/zsh) and fish wrapper variants are updated.
- The wrapper does NOT call `gcm config get` at runtime and does NOT consult any environment variable. See ADR 0004 for rationale.

## Acceptance criteria

- [ ] `gcm config set project-opener "code --new-window"` persists the value and exits 0
- [ ] `gcm config set project-opener ...` emits a one-line stderr hint pointing at `eval "$(gcm shell-init)"`
- [ ] `gcm config show` lists `project-opener` with the configured value, or with default annotation when unset
- [ ] No PATH validation is performed on the configured command
- [ ] After setting `project-opener` and re-running `eval "$(gcm shell-init)"`, `gcm open` followed by selection cd's into the repository *and* launches the configured command with `.` as its argument
- [ ] With `project-opener` unset, `gcm shell-init` emits a wrapper whose `open` branch is cd-only (regression check against 0015)
- [ ] The wrapper text contains the resolved opener as a literal string — no `gcm config get` call, no env-var lookups
- [ ] Fish-shell wrapper variant has equivalent behaviour
- [ ] Unit tests cover: opener baked into POSIX wrapper, opener baked into fish wrapper, unset opener produces cd-only wrapper, hint emitted only on `project-opener` set (not on other config keys)
- [ ] No change to `gcm clone` behaviour or its wrapper branch

## Blocked by

- 0015-gcm-open-command.md
