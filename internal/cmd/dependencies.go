package cmd

import (
	"io"
	"os/exec"

	"git-clone-manager/internal/configstore"
	"git-clone-manager/internal/gitrunner"
	"git-clone-manager/internal/statuspipeline"
)

type Dependencies struct {
	LoadEffectiveCloneConfig  func() (configstore.EffectiveConfig, error)
	NewGitRunner              func() gitrunner.Runner
	LoadEffectiveStatusConfig func() (configstore.EffectiveConfig, error)
	NewStatusCollector        func() statusCollector
	LoadEffectiveOpenConfig   func() (configstore.EffectiveConfig, error)
	OpenFZFAvailable          func() bool
	RunOpenFZF                func([]string, string, string) (string, int, error)
	OpenWriterIsTTY           func(any) bool
	LoadEffectiveShellConfig  func() (configstore.EffectiveConfig, error)
}

func DefaultDependencies() Dependencies {
	return Dependencies{
		LoadEffectiveCloneConfig: func() (configstore.EffectiveConfig, error) {
			return configstore.New().Effective()
		},
		NewGitRunner: gitrunner.New,
		LoadEffectiveStatusConfig: func() (configstore.EffectiveConfig, error) {
			return configstore.New().Effective()
		},
		NewStatusCollector: func() statusCollector {
			return statuspipeline.New(gitrunner.New())
		},
		LoadEffectiveOpenConfig: func() (configstore.EffectiveConfig, error) {
			return configstore.New().Effective()
		},
		OpenFZFAvailable: func() bool {
			_, err := exec.LookPath("fzf")
			return err == nil
		},
		RunOpenFZF: runFZF,
		OpenWriterIsTTY: func(writer any) bool {
			ioWriter, ok := writer.(io.Writer)
			return ok && writerIsTTY(ioWriter)
		},
		LoadEffectiveShellConfig: func() (configstore.EffectiveConfig, error) {
			return configstore.New().Effective()
		},
	}
}
