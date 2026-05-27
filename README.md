# git-clone-manager

`git-clone-manager` (`gcm`) is a zero-config CLI for cloning, organising, and monitoring many local git repositories.

It derives clone paths from repository URLs, keeps repositories under a predictable clone root, and scans that clone root to show which repositories are behind, dirty, or on non-default branches.

## Features

- Clone a repository by URL into a stable derived path:
  `${clone_root}/${hostname}/${path_prefix}/${repo_name}`
- Pass clone URLs to `git` as-is, so existing SSH keys and credential helpers continue to work.
- Scan every git repository under the clone root, even repositories not created by `gcm`.
- Show a status table with current branch, commits behind, dirty file count, and status badges.
- Fetch remotes by default during status collection, with `--no-fetch` for local-only checks.
- Filter status output to repositories on non-default branches with `--non-default`.
- Configure the clone root only when needed; the default is `~/src`.
- Install shell integration so `gcm clone <url>` leaves your shell in the cloned repository.

## Installation

Requirements:

- Go 1.25 or newer
- Git available on `PATH`
- Optional: [`fzf`](https://github.com/junegunn/fzf) available on `PATH` for `gcm open`

Build a local binary:

```sh
make build
```

Install with Go:

```sh
make install
```

Or run directly from source:

```sh
go run ./cmd/gcm --help
```

## Quick Start

Clone a repository:

```sh
gcm clone git@github.com:nWave-ai/aihero.git
```

By default, this clones to a path under `~/src` that mirrors the URL hierarchy, such as:

```text
~/src/github.com/nWave-ai/aihero
```

Show the status of repositories under the clone root:

```sh
gcm status
```

Install shell integration for clone auto-cd:

```sh
gcm shell-init --install
```

After restarting your shell, `gcm clone <url>` changes into the cloned repository when the clone succeeds.

## Usage Examples

See the rendered CLI examples and screen states on GitHub Pages:

[cli screens](https://stefaneg.github.io/gcm-aihero-variant/cli-surface/actual-cli-screens.html)

Common commands:

```sh
gcm clone https://github.com/example/org/repo.git
gcm status
gcm status --no-fetch
gcm status --non-default
gcm config show
gcm config set clone-root ~/src/work
gcm shell-init
gcm shell-init --install
```

## Configuration

`gcm` works without a config file. When no clone root is configured, it uses:

```text
~/src
```

Set a custom clone root:

```sh
gcm config set clone-root ~/src/work
```

Show the effective configuration:

```sh
gcm config show
```

The `GCM_CONFIG` environment variable can point to an alternate config file:

```sh
GCM_CONFIG=/tmp/gcm.yaml gcm config show
```

## Development Setup

Clone this repository and install dependencies:

```sh
git clone git@github.com:stefaneg/gcm-aihero-variant.git
cd gcm-aihero-variant
go mod download
```

Build the CLI:

```sh
make build
```

Run the CLI from source:

```sh
go run ./cmd/gcm --help
```

Useful project commands:

```sh
make build
make test
make run
make clean
```

## Testing

Run the test suite:

```sh
make test
```

The tests cover URL parsing, configuration storage, repository walking, git command execution boundaries, status collection, status formatting, shell initialization, and command behavior.

## Contributing

Issues are tracked as local markdown files under `docs/issues/`. Product requirements live under `docs/prd/`, and domain terminology is kept in `CONTEXT.md`.

Before making changes:

```sh
go test ./...
```

When adding user-facing CLI behavior, update the relevant command docs under `docs/cli-surface/` and keep stdout, stderr, and exit-code behavior aligned with `docs/specs/output-spec.md`.

## License

This project is licensed under the [MIT License](LICENSE).

## Troubleshooting

### `gcm clone` fails because the destination exists

`gcm clone` is idempotent only when the destination is already a git repository with a matching `origin` URL. If the derived path exists but is not a matching repository, move or remove that directory and run the command again.

### `gcm status` is slow

`gcm status` fetches remotes by default so behind counts are current. Use local git state only:

```sh
gcm status --no-fetch
```

### `gcm status` shows `[default-unknown]`

The repository's remote does not advertise `refs/remotes/origin/HEAD`. `gcm` does not guess a default branch name, so it marks the default branch as unknown.

### Shell integration does not change directories

Install the shell wrapper and restart your shell:

```sh
gcm shell-init --install
```

You can also evaluate it manually:

```sh
eval "$(gcm shell-init)"
```
