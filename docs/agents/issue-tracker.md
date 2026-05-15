# Issue tracker: Local Markdown

Issues and PRDs for this repo live as markdown files under `docs/`.

## Conventions

- Issues live under `docs/issues/` as `NNNN-<slug>.md`, numbered sequentially from `0001`
- PRDs live under `docs/prd/` as `<slug>.md`
- Triage state is recorded as a `Status:` line near the top of each issue file (see `triage-labels.md` for the role strings)
- Comments and conversation history append to the bottom of each file under a `## Comments` heading
- Issues reference blockers by filename (e.g. `0003-config-store-and-config-commands.md`)

## When a skill says "publish to the issue tracker"

Create a new file under `docs/issues/` (for issues) or `docs/prd/` (for PRDs), following the naming convention above. Add a `Status: ready-for-agent` line near the top unless instructed otherwise.

## When a skill says "fetch the relevant ticket"

Read the file at `docs/issues/NNNN-<slug>.md`. The user will normally pass the issue number or path directly.

## When a skill says "close an issue"

Add `Status: wontfix` or `Status: done` at the top of the file. Do not delete the file.
