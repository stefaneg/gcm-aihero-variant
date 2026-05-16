package repositorywalker

import (
	"io/fs"
	"path/filepath"
	"slices"
)

func Walk(cloneRoot string) ([]string, error) {
	var repositories []string

	err := filepath.WalkDir(cloneRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() && entry.Name() == ".git" {
			repositoryPath := filepath.Dir(path)
			if repositoryPath != cloneRoot {
				repositories = append(repositories, repositoryPath)
			}
			return filepath.SkipDir
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	slices.Sort(repositories)
	return repositories, nil
}
