package cmd

import (
	"bytes"
	"testing"

	"git-clone-manager/internal/configstore"
	"git-clone-manager/internal/gitrunner"
	"git-clone-manager/internal/gitrunnertest"
)

func TestClonePassesURLToGitAsIs(t *testing.T) {
	fakeRunner := gitrunnertest.New()

	originalLoadConfig := loadEffectiveCloneConfig
	originalNewGitRunner := newGitRunner
	t.Cleanup(func() {
		loadEffectiveCloneConfig = originalLoadConfig
		newGitRunner = originalNewGitRunner
	})

	loadEffectiveCloneConfig = func() (configstore.EffectiveConfig, error) {
		return configstore.EffectiveConfig{CloneRoot: t.TempDir()}, nil
	}
	newGitRunner = func() gitrunner.Runner {
		return fakeRunner
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	rawURL := "custom+git://example.com/deep/group/repo.git"
	exitCode := Execute([]string{"clone", rawURL}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("Execute exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	cloneCalls := fakeRunner.CloneCalls()
	if len(cloneCalls) != 1 {
		t.Fatalf("CloneCalls length = %d, want 1", len(cloneCalls))
	}

	if cloneCalls[0].URL != rawURL {
		t.Fatalf("clone URL = %q, want %q", cloneCalls[0].URL, rawURL)
	}
}
