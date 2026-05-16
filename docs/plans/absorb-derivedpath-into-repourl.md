# Absorb derivedpath into repourl.Parts

## Problem

`clone.go` parses a URL into a `repourl.Parts` struct and immediately destructures it
to pass four separate strings to `derivedpath.Derive`:

```go
// internal/cmd/clone.go:46
destinationPath := derivedpath.Derive(cloneRoot, parts.Hostname, parts.PathPrefix, parts.RepositoryName)
```

The struct exists as a named unit but cannot hold itself together across the seam.
`derivedpath` is 12 LOC wrapping `filepath.Join` — the deletion test shows it earns
no leverage of its own; inlining its logic as a method on `Parts` concentrates the
concept rather than scattering it.

## Fix

Add `DerivedPath(cloneRoot string) string` as a method on `repourl.Parts`. Delete the
`derivedpath` package entirely.

## Steps

### 1. Add method to `internal/repourl/parse.go`

Add `"path/filepath"` to imports, then add after `parseParts`:

```go
func (p Parts) DerivedPath(cloneRoot string) string {
    if p.PathPrefix == "" {
        return filepath.Join(cloneRoot, p.Hostname, p.RepositoryName)
    }
    return filepath.Join(cloneRoot, p.Hostname, p.PathPrefix, p.RepositoryName)
}
```

### 2. Migrate tests to `internal/repourl/parse_test.go`

The two cases in `derivedpath/derive_test.go` (`with path prefix`, `without path prefix`)
become a `TestDerivedPath` table test in the repourl test file, calling
`parts.DerivedPath(cloneRoot)`.

### 3. Update `internal/cmd/clone.go`

Line 46 changes from:

```go
destinationPath := derivedpath.Derive(cloneRoot, parts.Hostname, parts.PathPrefix, parts.RepositoryName)
```

to:

```go
destinationPath := parts.DerivedPath(cloneRoot)
```

Drop the `"git-clone-manager/internal/derivedpath"` import.

### 4. Delete `internal/derivedpath/`

Remove both `derive.go` and `derive_test.go`. The package has no other callers.

## Outcome

- **Locality**: the Derived Path computation lives next to the data it computes from.
- **Leverage**: callers get one call with no destructuring and one fewer import.
- The domain concept (Derived Path, defined in `CONTEXT.md`) is preserved as a method name.
