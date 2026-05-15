# Protocol passed as-is to git

gcm passes the repository URL to git without rewriting the scheme. If the user provides `https://github.com/org/repo`, git clones over HTTPS. If they provide `git@github.com:org/repo`, git clones over SSH. gcm never infers or overrides the protocol.

The original design (constraint 5 in the user stories) said protocol was always applied from the hosting platform config and never inferred from the URL. We reversed this when hosting platform entries became optional: without a mandatory config step, there is no authoritative source for a protocol preference. Passing the URL as-is is the correct default — it matches what `git clone` itself does and gives the user full control without requiring configuration.

When a hosting platform entry exists and specifies a protocol override, that override applies (post-v1). In v1, the URL scheme is the protocol.
