# 0026 — README: document `fzf` as a required dependency for `gcm open`

Status: ready-for-agent

## What to build

`README.md`'s Requirements section (around line 21) lists Go and Git as runtime requirements but does not mention `fzf`. After 0015 lands, `gcm open` hard-depends on `fzf` being on the user's `PATH` and exits with an actionable error when it's missing. A user reading the README to install gcm should know up front that `fzf` is needed if they want `gcm open` — not discover it the first time they try the command.

Update the Requirements section to list `fzf` as an *optional* dependency required only for `gcm open`. Keep Go and Git as required (they're needed for build / core commands). One line is enough — link to `https://github.com/junegunn/fzf` for install instructions rather than duplicating per-platform install steps.

If the README has a per-command section that documents `gcm open` (or gets one as part of 0015), the `fzf` requirement should also be mentioned there so a reader scanning that section in isolation sees it.

## Acceptance criteria

- [ ] README's Requirements section lists `fzf` as an optional dependency required for `gcm open`, with a link to the fzf project
- [ ] If `gcm open` has its own README section, that section also mentions the `fzf` requirement
- [ ] Go and Git remain listed as required (no regression to existing requirements)
- [ ] No code changes — README only

## Blocked by

- 0015-gcm-open-command.md
