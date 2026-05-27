# 0020 — workerpool: recover panics in workers and surface as `Result.Err`

Status: done

## What to build

`internal/workerpool` workers run `work(item)` without a `defer recover()`. A panic in any worker tears down the whole process via Go's default panic behaviour. PRD story #26 ("under 10 seconds for 200 repositories") implicitly relies on the pool being robust — a single bad repository should not abort the run.

Wrap each work invocation in a `defer recover()`. On recover, construct a `Result.Err` whose message includes the panic value (and ideally the call site or a short stack trace, runtime/debug.Stack is fine). The pool continues to drain remaining work. Downstream consumers (`statuspipeline`, see 0019) treat this as a hard per-item error and surface a row with an error badge.

This is intentionally narrow: do not change the pool's public API, the channel topology, or the worker count semantics. Just contain the panic.

## Acceptance criteria

- [ ] A worker that panics on one input does not crash the process; remaining inputs are processed
- [ ] The panicking input produces a `Result` whose `Err` is non-nil and whose message conveys it came from a panic (e.g. `worker panic: <value>`), with enough context for a user to identify which repository panicked
- [ ] No change to the pool's exported signature or behaviour for non-panicking workers
- [ ] Unit tests cover: single panicking worker among many succeeding ones; panic message is preserved in `Result.Err`

## Blocked by

- None - can start immediately
