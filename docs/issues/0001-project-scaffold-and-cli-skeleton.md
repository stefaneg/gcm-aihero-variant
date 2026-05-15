# 0001 — Go project scaffold and CLI skeleton

## What to build

Initialise the Go module and wire up a Cobra CLI with the full v1 command skeleton. No business logic yet — just the command tree, help text, flag declarations, and exit code infrastructure. After this slice, `gcm --help`, `gcm clone --help`, `gcm status --help`, `gcm config --help`, `gcm config set --help`, and `gcm config show --help` all work and display correct usage.

The binary entry point is `cmd/gcm/main.go`. Commands live under `internal/cmd/`. Each command is a stub that prints "not yet implemented" and exits 0.

## Acceptance criteria

- [ ] `go build ./...` produces a `gcm` binary with no errors
- [ ] `gcm --help` lists clone, status, and config subcommands
- [ ] `gcm status --help` documents `--no-fetch` and `--non-default` flags
- [ ] `gcm config set --help` documents the `clone-root` subcommand
- [ ] `gcm config show --help` renders correctly
- [ ] Exit code infrastructure supports 0 (success), 1 (general error), 2 (usage/config error)
- [ ] All commands stub out with "not yet implemented" rather than panicking

## Blocked by

None — can start immediately.
