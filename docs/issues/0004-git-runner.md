# 0004 — Git runner

## What to build

A Git Runner module that is the single point of contact between gcm and the git binary. No other module shells out to git. All git subprocess execution lives here.

The Git Runner exposes an interface (not a concrete struct) so that tests in other modules can inject a fake runner without touching the filesystem or spawning processes. The interface covers the operations needed by v1: clone, fetch, status (dirty file count), rev-list (commits behind), and symbolic-ref (default branch detection).

Credentials are never passed through the Git Runner — they flow through git's own credential system (SSH keys, credential helpers). The runner simply invokes git with the arguments it is given and returns structured output or a typed error.

## Acceptance criteria

- [ ] Git Runner exposes an interface that other modules depend on (not a concrete type)
- [ ] Implements: `Clone(url, destPath string)`, `Fetch(repoPath string)`, `DirtyCount(repoPath string) (int, error)`, `CommitsBehind(repoPath string) (int, error)`, `DefaultBranch(repoPath string) (string, error)`
- [ ] `DefaultBranch` reads `refs/remotes/origin/HEAD` and returns an error when not set (callers handle the fallback to `main`)
- [ ] Typed errors distinguish: git not found, repository not found, network/fetch failure, no remote configured
- [ ] No credentials, tokens, or API keys are passed as arguments to any git subprocess
- [ ] A fake implementation of the interface is provided in a test helper package for use by Status Collector and other callers
- [ ] Unit tests for the real runner invoke git against a temp repository created during the test

## Blocked by

- 0001 — project scaffold and CLI skeleton
