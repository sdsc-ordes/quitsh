package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
)

// Register registers the build runner in the factory.
func RegisterBuild(
	buildSettings config.IBuildSettings,
	factory factory.IFactory,
	registerKey bool,
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

	if registerKey {
		e = factory.RegisterToKey(runner.NewRegisterKey("build", "go"), GoBuildRunnerID)
		err = errors.Combine(err, e)
	}

	return err
}

func RegisterTest(
	testSettings config.ITestSettings,
	factory factory.IFactory,
	registerKey bool,
) (err error) {
	// Register Go test/test-bin runner.
	log.Trace("Register runner.", "id", GoTestRunnerID)
	e := factory.Register(
		GoTestRunnerID,
		runner.RunnerData{
			Creator: func(runnerConfig any) (runner.IRunner, error) {
				return NewGoTestRunner(runnerConfig, testSettings)
			},
			RunnerConfigUnmarshal: UnmarshalBuildConfig,
			DefaultToolchain:      "go",
		})
	err = errors.Combine(err, e)

	if registerKey {
		e = factory.RegisterToKey(runner.NewRegisterKey("test", "go"), GoTestRunnerID)
		err = errors.Combine(err, e)
	}

	log.Trace("Register runner.", "id", GoTestBinRunnerID)
	e = factory.Register(
		GoTestBinRunnerID,
		runner.RunnerData{
			Creator: func(runnerConfig any) (runner.IRunner, error) {
				return NewGoTestBinRunner(runnerConfig, testSettings)
			},
			RunnerConfigUnmarshal: UnmarshalTestBinConfig,
			DefaultToolchain:      "go",
		})
	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey("test", "go-bin"), GoTestBinRunnerID)
	err = errors.Combine(err, e)

	return err
}
