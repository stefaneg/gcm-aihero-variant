# Zero-config clone

`gcm clone` requires no configuration file and no hosting platform entry to run. The only input is the URL; the only prerequisite is that the clone root exists (defaulting to `~/src`, created with a warning if absent). git receives the URL as-is and handles credentials via its own credential system.

The original design required a hosting platform entry before cloning — for protocol selection and credential routing. We reversed this because it created a bootstrapping problem: a user who just installed gcm had to configure it before they could use its primary command. Removing the prerequisite means the tool is useful immediately, and hosting platform entries remain an opt-in enhancement for protocol override and API access (post-v1).

## Consequences

`gcm clone` cannot enforce a protocol preference without a hosting platform entry. Users who want SSH for a host that they typically reach via HTTPS must configure a hosting platform entry. This is a deliberate trade-off: the common case (clone this URL) is zero-friction; the power case (always use SSH for this host) requires explicit configuration.
