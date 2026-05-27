# 0021 — gitrunner: context cancellation + locale-independent stderr classification

Status: done

## What to build

Two robustness fixes in `internal/gitrunner`. Both touch `runGitCommand` and ship together.

**Context cancellation / per-call timeout.** `runGitCommand` uses `exec.Command` with no context, so a hung `git fetch` (unreachable remote that doesn't time out at the network layer, SSH prompting for input, etc.) blocks the calling worker indefinitely. PRD story #26 promises "under 10 seconds for 200 repositories." Plumb a `context.Context` through the `gitrunner` public API (or accept a per-call timeout if that's a cleaner fit) and use `exec.CommandContext`. Callers in `statuscollector` / `statuspipeline` pass a bounded context per repository.

The default timeout value is an implementer call — somewhere in the 30–60s range is plausible; expose it as a constant or option rather than burying it. A timed-out command produces a classified error (a new `TimeoutError`, or `NetworkError` reused if the implementer judges the user-visible behaviour matches) so the pipeline can render a `[fetch-failed]`-style row instead of aborting.

**Locale-independent stderr classification.** `classifyError` matches English keywords like `"could not resolve host"` against git's stderr. Under a non-C locale (`LANG=de_DE.UTF-8`) git emits localised messages and every soft-error case (network, no-remote, origin-HEAD-unset) falls through to "unknown hard error," which today aborts the batch and (after 0019) renders rows as hard errors — visibly degraded UX. Force the C locale on the spawned process by setting `LC_ALL=C` (and `LANG=C` for belt-and-braces) in the command's environment. This is a one-line env addition at the `exec.Cmd` construction site.

## Acceptance criteria

- [ ] `runGitCommand` accepts a context (or per-call timeout) and uses `exec.CommandContext`
- [ ] A command that exceeds the timeout produces a classified error that the pipeline treats as a soft per-repo failure, not a batch abort
- [ ] All spawned git commands run with `LC_ALL=C` (and `LANG=C`) in their environment
- [ ] A test simulating localised git stderr (or directly feeding non-English stderr to `classifyError`) confirms classification still succeeds because the runtime env forces C
- [ ] Unit tests cover: context cancellation produces the expected error; timeout produces the expected error; env-var override at exec site is present
- [ ] Existing gitrunner tests pass; `gitrunnertest` fake is updated if the public signature changes

## Blocked by

- None - can start immediately
