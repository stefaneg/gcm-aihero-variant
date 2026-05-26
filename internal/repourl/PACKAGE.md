# `git-clone-manager/internal/repourl` — Repository URL parsing and derived paths

## Glossary

Domain vocabulary lives in [`../../CONTEXT.md`](../../CONTEXT.md).

## Package scope

This package belongs to the clone workflow's URL interpretation layer.

## Core concept owned

This package is the canonical home for extracting repository URL parts and computing the derived path.

## Responsibilities

- Owns: parsing HTTPS-style URLs and scp-style `git@host:path` URLs into `Parts`.
- Owns: `ParseError` with the raw URL and reason for malformed input.
- Owns: stripping a trailing `.git` suffix from the repository path before path derivation.
- Owns: computing derived paths as `<cloneRoot>/<hostname>/<pathPrefix>/<repositoryName>`.
- Does **not** own: cloning or protocol rewriting; `internal/gitrunner` runs git with the original URL.
- Does **not** own: clone-root defaults or persistence; `internal/configstore` owns configuration.
- Does **not** own: filesystem preparation for the derived path; `internal/cmd` owns command execution flow.

## Upstream (this package depends on)

- None.

## Downstream (consumers of this package)

- `git-clone-manager/internal/cmd` — computes clone destinations for `gcm clone`.
- `git-clone-manager/cmd/gcm` tests — verifies executable-level derived path behaviour.

## Invariants & conventions

- The raw URL is parsed only to derive local path parts; the original URL remains the value passed to git.
- Missing hostname, repository path, or repository name returns `*ParseError`.
- `PathPrefix` is slash-separated URL path vocabulary; `DerivedPath` converts it with `filepath.Join`.
- Platform terms such as org, owner, namespace, and group do not appear in this package's API.

## When developing in this package

- [ ] Did any new URL handling keep path-prefix vocabulary platform-agnostic and leave protocol choice with the
  user's raw URL?

## See also

- [`../../CONTEXT.md`](../../CONTEXT.md) — derived-path and path-prefix vocabulary.
- [`../../docs/adr/0003-protocol-passed-as-is.md`](../../docs/adr/0003-protocol-passed-as-is.md) — URL protocol is
  not rewritten before cloning.

## Clean-concept rating — PASS

This package owns a single concept cleanly: converting repository URLs into stable path parts. The smell test passes
because parsing and path derivation are one lifecycle, with no package-level mutable state and no self-admission
comments.
