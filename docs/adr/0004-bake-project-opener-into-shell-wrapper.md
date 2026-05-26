# Bake project-opener into the shell wrapper at shell-init time

The `gcm open` shell wrapper, emitted by `gcm shell-init`, contains the resolved `project-opener` command as a literal string. The wrapper does not consult `gcm config` or any environment variable at invocation time. Changing the configured opener requires re-running `eval "$(gcm shell-init)"` (or re-sourcing the rc file) before the change takes effect.

The alternative was a wrapper that resolves the opener at every invocation — either by calling `gcm config get project-opener` from the shell on each call, or by emitting a fallback chain (`${GCM_OPENER:-...}`) into the wrapper text. Runtime resolution would mean `gcm config set project-opener <cmd>` takes effect immediately and supports per-shell `GCM_OPENER=... gcm open` overrides without re-emitting the wrapper. We chose baking instead, for three reasons:

1. **Wrapper symmetry with `gcm clone`.** The existing `clone` branch of the wrapper is a thin shim: capture stdout, cd. Adding `gcm config get` calls or env-var fallback logic only to the `open` branch would make the two branches diverge in shape and complexity. A single literal command keeps both branches the same shape.
2. **No surprise interleaving.** A runtime `gcm config get` call inside the wrapper would spawn a `gcm` process before fzf runs, on every invocation. Baking avoids the extra process and the failure modes that come with it (config file race, transient PATH issues).
3. **The opener rarely changes.** A developer picks an IDE once and uses it for months. Optimising the wrapper for the change-frequency of the config value, rather than the invocation frequency of `gcm open`, would be backwards.

The cost is that `gcm config set project-opener <cmd>` is the only config key whose effect is not immediate. To make this discoverable, `gcm config set project-opener` prints a one-line stderr hint reminding the user to re-run `shell-init`. When `project-opener` is unset at shell-init time, the emitted wrapper's `open` branch only cd's into the selected repository and launches nothing.

This decision is scoped to `project-opener`. `clone-root` and other config keys are read at runtime by the Go binary itself, so they take effect immediately and require no re-bake.
