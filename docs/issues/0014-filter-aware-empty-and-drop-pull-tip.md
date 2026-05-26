# 0014 — Filter-aware empty `--non-default`; drop `gcm pull` tip

Status: ready-for-agent

## What to build

Two small Status Formatter changes that resolve PRD-flagged ambiguities, packaged together because they both touch the table's footer (summary + tips).

**Filter-aware empty `--non-default`**. When `gcm status --non-default` matches zero repositories against a non-empty clone root, the table is distinguishable from "no repositories under clone root":

- A filter-aware message in place of table rows, naming the filter (e.g., `No repositories on non-default branches.`)
- A summary line that shows both the total repository count and the non-default count, so the user can see the filter ran against a real set (e.g., `42 repositories, 0 non-default.`)
- The `--non-default` tip is suppressed in this state (telling the user to run the filter that just yielded zero rows is noise)

The fully-empty clone root case is unchanged from issue 0009.

**Drop the `gcm pull` tip**. `gcm pull` is post-v1. Surfacing it as a tip points the user at a "command not found." Until `gcm pull` ships, the only tip rendered is the one pointing to `gcm status --non-default`.

When `gcm pull` lands in a later release, the tip can be reinstated; tip wording is `Convenience`, not `Contract`, per the project output spec.

## Acceptance criteria

- [ ] `gcm status --non-default` against a clone root with repositories, none on non-default branches, prints the filter-aware message and the dual-count summary
- [ ] In that state, no `--non-default` tip line is rendered
- [ ] `gcm status --non-default` against an empty clone root behaves exactly as `gcm status` against an empty clone root (no change)
- [ ] `gcm status` (no flags, populated clone root) renders the `gcm status --non-default` tip but does **not** render any `gcm pull` tip
- [ ] Unit tests cover: filter-empty case, fully-empty case, populated case with and without non-default rows
- [ ] Existing tip-related assertions referring to `gcm pull` are removed or rewritten

## Blocked by

- None
