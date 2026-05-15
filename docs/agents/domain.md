# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

## Before exploring, read these

- **`CONTEXT.md`** at the repo root — the project glossary. Use its terms; avoid the synonyms listed under each entry.
- **`docs/adr/`** — read ADRs that touch the area you're about to work in before making implementation decisions.

If either doesn't exist yet, proceed silently. The `/grill-with-docs` skill creates them lazily as terms and decisions get resolved.

## File structure

Single-context repo:

```
/
├── CONTEXT.md
├── docs/
│   └── adr/
│       ├── 0001-filesystem-walk-for-repository-discovery.md
│       ├── 0002-zero-config-clone.md
│       └── 0003-protocol-passed-as-is.md
```

## Use the glossary's vocabulary

When your output names a domain concept (in an issue title, a refactor proposal, a hypothesis, a test name), use the term as defined in `CONTEXT.md`. The glossary's avoid-lists are as important as the definitions — don't drift to the listed synonyms.

Key terms for this project: **Repository**, **Clone Root**, **Derived Path**, **Path Prefix**, **Default Branch**, **Behind**, **Dirty**, **Hosting Platform**, **Discover**, **Status Table**.

## Flag ADR conflicts

If your output contradicts an existing ADR, surface it explicitly rather than silently overriding:

> _Contradicts ADR-0002 (zero-config clone) — but worth reopening because…_
