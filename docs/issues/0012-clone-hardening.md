# 0012 — `gcm clone` hardening: origin-mismatch, empty-dir, partial-cleanup

Status: ready-for-agent

## What to build

Three independent refinements to `gcm clone`'s existing-destination and failure handling, all motivated by the shell-wrapper UX (issue 0011): the wrapper must never cd into the wrong repo, must not refuse benign pre-existing directories, and must not leave broken state behind on failure.

**Origin-mismatch error path.** When the destination is an existing git repository whose `origin` URL does **not** match the URL the user asked to clone, this is an error (not an idempotent skip). Emit a stderr error following the project error style — naming both the destination path and the conflicting origin URL — and exit 1. No stdout. This complements issue 0010, which handles the matching-origin case as a silent re-emit of the path.

**Empty pre-existing destination accepted.** A pre-existing directory at the derived path that contains zero entries is accepted as a valid clone target, matching `git clone`'s own behaviour. Any single entry — including dotfiles such as `.DS_Store` — counts as non-empty and continues to be blocked. The empty-directory path must produce the same stdout/stderr split as a fresh clone (path on stdout, progress on stderr, exit 0).

**Partial-clone cleanup on failure.** When `git clone` fails part-way through, gcm removes the directories it itself created during the run; a pre-existing directory the user made is left in place. The Git Runner / clone command layer must track which directories it created so it can clean up only its own writes. On failure, exit code and error message are unchanged from current behaviour — only the filesystem state is restored.

## Acceptance criteria

- [ ] Cloning over an existing git repo with a mismatched origin emits a stderr error naming the destination path and the existing origin URL, no stdout, exit 1
- [ ] Cloning into a pre-existing empty directory succeeds; the directory is reused, the destination path is emitted on stdout, exit 0
- [ ] Cloning into a pre-existing directory containing only `.DS_Store` is blocked with the existing non-git-destination error
- [ ] A simulated mid-clone failure (e.g., killing `git clone`, or pointing at an unreachable URL) leaves no new directories or files under the clone root that gcm itself created
- [ ] A simulated mid-clone failure into a pre-existing empty directory leaves the directory itself intact (gcm did not create it; gcm does not delete it)
- [ ] Integration tests cover all three scenarios end-to-end using a local bare repository as the remote

## Blocked by

- 0010 — `gcm clone` adopts stdout-as-result contract
