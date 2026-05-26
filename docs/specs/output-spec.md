# CLI Output Spec

This document defines the project-wide output contract for `gcm` commands.

## Streams

- stdout is reserved for command results that a human can read or a script can pipe onward.
- stderr is reserved for diagnostics, warnings, errors, and progress messages.
- Progress messages (e.g. "Cloning to ...", "Done.") are diagnostics and belong on stderr, never stdout.
- For action commands that change state, the **result** on stdout is the identifier of what was produced or affected — typically a path. The result is emitted as a single bare line (no `S|` prefix, no surrounding decoration) so that callers can capture it with `dest=$(gcm <subcommand> ...)`. This contract is what enables shell-function wrappers (see `gcm shell-init`) to compose with action commands without parsing or special modes.
- Commands must not interleave diagnostics into stdout.
- Success output should be plain text. `gcm` does not support JSON output in v1.

## Formats

- The only supported output format is the default plain-text format.
- `--format`, `--output`, and equivalent structured-output flags are not part of the v1 contract.
- Commands should prefer stable, simple text:
  - one-line result messages for single actions;
  - aligned rows for multi-repository status output;
  - key/value lines for small configuration output.

## Errors

- Errors are written to stderr.
- Error messages use a one-part form:

```text
Error: <noun> "<identifier>": <reason>
```

- Every error message must name the offending command, flag, argument, path, URL, config key, or repository root literally.
- There is no hint system. Commands must not emit separate `Hint:` lines.
- When remediation is needed, include it in the reason sentence itself.

## Exit Codes

| Code | Meaning |
|---:|---|
| 0 | success |
| 1 | general runtime failure, including partial batch failure |
| 2 | usage or configuration error |

- Usage errors include missing arguments, unknown commands, unknown flags, and invalid flag values.
- Partial batch failure means the command may still print partial results before exiting non-zero.
- Exit codes are the stable contract; exact human wording is convenience text.

## Color and TTY

- Text badges must always be present as text.
- Color is a progressive enhancement only.
- Color is emitted only when the relevant stream is a TTY and `NO_COLOR` is unset.
- stdout and stderr TTY checks are independent.
- v1 does not provide `--color`, `--no-color`, or equivalent color-control flags.

## State-Changing Commands

- v1 does not support dry-run flags.
- Destructive commands should prompt before destructive action unless they are explicitly designed as non-destructive or idempotent.
- `gcm clone` is idempotent: rerunning a clone against an existing git repository with a matching `origin` URL re-emits the destination path on stdout and exits 0 with no progress noise on stderr. A pre-existing git repository whose `origin` does not match the requested URL is treated as an error (exit 1), not as an idempotent skip.
- `gcm config set clone-root` writes configuration immediately when invoked with a path.

## Help and Discoverability

- No-argument root invocation prints top-level help and exits 0.
- Group-only invocation prints group help and exits 0.
- Missing required input prints an error to stderr and exits 2.
- Cobra's `help` and `completion` commands are part of the v1 surface.

## Stability

| Element | Stability |
|---|---|
| Command names and tree position | Contract |
| Long flag names | Contract |
| Short flag names | Contract |
| Exit codes | Contract |
| Plain stdout shape | Convenience |
| Error wording | Convenience |
| stderr diagnostic wording | Convenience |
