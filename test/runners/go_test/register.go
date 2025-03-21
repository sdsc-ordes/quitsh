//go:build test

package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
	settings "github.com/sdsc-ordes/quitsh/test/runners/settings_test"
)

// Register registers the runners in the factory.
func Register(
	buildSettings *settings.BuildSettings,
	factory factory.IFactory,
) (err error) {
	// Register Go build runner.
	log.Trace("Register runner.", "id", GoBuildRunnerID)
	e := factory.Register(
		GoBuildRunnerID,
		runner.RunnerData{
			Creator: func(runnerConfig any) (runner.IRunner, error) {
				return NewGoBuildRunner(runnerConfig, buildSettings)
			},
			RunnerConfigUnmarshal: UnmarshalBuildConfig,
			DefaultToolchain:      "go",
		})
	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey("build", "go-custom"), GoBuildRunnerID)
	err = errors.Combine(err, e)

	return err
}
