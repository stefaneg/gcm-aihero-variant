# 0023 — cmd/execute: replace substring "unknown command" match with Cobra sentinel

Status: ready-for-agent

## What to build

`internal/cmd/execute.go` classifies "unknown command" errors via `strings.HasPrefix(message, "unknown command ")`. This is fragile: a localised Cobra build, a Cobra version that tweaks its wording (e.g. "unknown subcommand"), or an `ErrSubCommandRequired`-style branch would all defeat the match and re-classify the error as "general" — landing the user on exit 1 instead of the PRD-mandated exit 2 for usage errors (story #39).

Replace the string match with Cobra's exposed error sentinel or sentinel-equivalent (`errors.Is(err, cobra.ErrUnknownCommand)` if available, or whatever the installed Cobra version exports). If no suitable sentinel exists, document the version constraint and consider a Cobra upgrade as part of this change.

This is a small, targeted fix — no other classification logic in `execute.go` changes.

## Acceptance criteria

- [ ] `gcm <not-a-command>` exits with code 2 (usage error per PRD #39)
- [ ] The "unknown command" path is recognised via a typed/sentinel check, not a substring match on the error message
- [ ] A unit test asserts the exit code for an unknown subcommand, not just the message
- [ ] Other branches in `execute.go` are untouched
- [ ] If Cobra is upgraded as part of this change, the upgrade is the smallest version bump that exposes the sentinel

## Blocked by

- None - can start immediately
