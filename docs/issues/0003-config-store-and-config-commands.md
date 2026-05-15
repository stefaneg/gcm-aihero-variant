# 0003 — Config store and config commands

## What to build

A Config Store module and the two v1 config commands wired end-to-end.

**Config Store**: reads and writes `~/.config/gcm/config.yaml` (or the path in `$GCM_CONFIG`). Exposes an effective-config view that merges on-disk values over hard-coded defaults (clone root defaults to `~/src`). Never auto-creates the config file on a read — the file only appears when `gcm config set` is explicitly called. The only module in the codebase that touches the config file.

**`gcm config show`**: prints the effective configuration. Defaults are annotated (e.g., `clone_root: ~/src  # default`) so the user can distinguish configured values from fallback values.

**`gcm config set clone-root <path>`**: writes the given path as `clone_root` in the config file. Creates the config file and its parent directory if they do not exist (mode 0600 for the file). Prints the path where config was saved.

## Acceptance criteria

- [ ] `gcm config show` with no config file prints `clone_root: ~/src  # default` and exits 0
- [ ] `gcm config show` with a config file prints the configured value without the `# default` annotation
- [ ] `gcm config set clone-root /custom/path` writes the value and confirms the saved path
- [ ] After `gcm config set clone-root /custom/path`, `gcm config show` reflects the new value
- [ ] `$GCM_CONFIG` env var overrides the default config file path for both read and write
- [ ] No config file is created by `gcm config show` (read-only operation)
- [ ] Config file is created with mode 0600
- [ ] `gcm config set` with a missing argument exits 2 with a usage error
- [ ] Unit tests use a temp directory as config home; no tests touch the real `~/.config`

## Blocked by

- 0001 — project scaffold and CLI skeleton
