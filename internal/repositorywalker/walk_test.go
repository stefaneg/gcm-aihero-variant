package repositorywalker_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"git-clone-manager/internal/repositorywalker"
)

func TestWalkReturnsDirectoriesContainingGitSubdirectory(t *testing.T) {
	cloneRoot := t.TempDir()

	firstRepository := filepath.Join(cloneRoot, "github.com", "acme", "api")
	secondRepository := filepath.Join(cloneRoot, "gitlab.com", "team", "worker")
	createGitRepositoryDirectory(t, firstRepository)
	createGitRepositoryDirectory(t, secondRepository)

	plainDirectory := filepath.Join(cloneRoot, "notes")
	if err := os.MkdirAll(plainDirectory, 0o755); err != nil {
		t.Fatalf("create plain directory: %v", err)
	}

	got, err := repositorywalker.Walk(cloneRoot)
	if err != nil {
		t.Fatalf("Walk returned error: %v", err)
	}

	want := []string{firstRepository, secondRepository}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Walk() = %#v, want %#v", got, want)
	}
}

func TestWalkReturnsEmptySliceForEmptyCloneRoot(t *testing.T) {
	cloneRoot := t.TempDir()

	got, err := repositorywalker.Walk(cloneRoot)
	if err != nil {
		t.Fatalf("Walk returned error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("Walk() returned %d repositories, want 0", len(got))
	}
}

func TestWalkReturnsErrorWhenCloneRootDoesNotExist(t *testing.T) {
	cloneRoot := filepath.Join(t.TempDir(), "missing")

	_, err := repositorywalker.Walk(cloneRoot)
	if err == nil {
		t.Fatal("Walk unexpectedly succeeded")
	}

	if !os.IsNotExist(err) {
		t.Fatalf("Walk error = %v, want not-exist error", err)
	}
}

func TestWalkDoesNotReturnCloneRootItselfWhenItIsGitRepository(t *testing.T) {
	cloneRoot := t.TempDir()
	createGitRepositoryDirectory(t, cloneRoot)

	got, err := repositorywalker.Walk(cloneRoot)
	if err != nil {
		t.Fatalf("Walk returned error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("Walk() = %#v, want empty slice", got)
	}
}

func createGitRepositoryDirectory(t *testing.T, repositoryPath string) {
	t.Helper()

	gitDirectory := filepath.Join(repositoryPath, ".git")
	if err := os.MkdirAll(gitDirectory, 0o755); err != nil {
		t.Fatalf("create git directory: %v", err)
	}
}
