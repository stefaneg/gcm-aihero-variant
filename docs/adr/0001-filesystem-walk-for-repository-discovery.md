# Filesystem walk for repository discovery

`gcm status` discovers repositories by walking the clone root at runtime, finding all `.git` directories, rather than maintaining a manifest of repositories it has cloned. Any git repository under the clone root is in scope, regardless of whether gcm created it.

We chose the walk because a manifest drifts: repositories moved, deleted, or cloned by other tools outside gcm would fall out of sync, requiring reconciliation logic. The walk is always accurate by definition. The performance cost is acceptable — the walk itself is fast; the bottleneck is the parallel git fetch, not directory traversal.

## Considered options

A registry (manifest file updated on every `gcm clone` and `gcm discover`) would make `gcm status` faster to initialise and would let gcm distinguish "managed" from "unmanaged" repositories. We rejected it because drift between the manifest and the filesystem is a correctness problem, not just a cosmetic one — a stale manifest causes silent omissions in status output.
