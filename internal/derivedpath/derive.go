package derivedpath

import "path/filepath"

func Derive(cloneRoot, hostname, pathPrefix, repositoryName string) string {
	if pathPrefix == "" {
		return filepath.Join(cloneRoot, hostname, repositoryName)
	}

	return filepath.Join(cloneRoot, hostname, pathPrefix, repositoryName)
}
