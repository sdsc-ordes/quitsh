//go:build test

package gorunner

import (
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	gox "github.com/sdsc-ordes/quitsh/pkg/exec/go"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	setts "github.com/sdsc-ordes/quitsh/test/runners/settings_test"
)

const GoBuildRunnerID = "integration-test::build-go"

type GoBuildRunner struct {
	runnerConfig *GoBuildConfig
	settings     *setts.BuildSettings
}

// NewGoBuildRunner constructs a new GoBuildRunner with its own config.
func NewGoBuildRunner(config any, settings *setts.BuildSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &GoBuildRunner{
		runnerConfig: common.Cast[*GoBuildConfig](config),
		settings:     settings,
	}, nil
}

func (*GoBuildRunner) ID() runner.RegisterID {
	return GoBuildRunnerID
}

func (r *GoBuildRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	comp := ctx.Component()

	log.Info("Hello from integration test Go runner.", "component", comp.Name())
	fs.AssertDirs(comp.OutBuildBinDir())

	log.Infof("OutputDir: %v", comp.OutDir())
	binDir := comp.OutBuildBinDir()

	goctx := gox.NewCtxBuilder().
		Cwd(comp.Root()).
		Env("GOBIN="+binDir,
			"GOWORK=off",
			"GOTOOLCHAIN=local",
		).
		Build()

	if r.settings.BuildType == common.BuildRelease {
		log.Info("Hurrey building release version")
	}

	log.Info("Run Go install.")
	cmd := []string{"install"}
	cmd = append(cmd, path.Join(comp.Root(), "..."))
	err := goctx.Check(cmd...)

	if err != nil {
		log.ErrorE(err, "Go install failed.")

		return err
	}

	return nil
}
