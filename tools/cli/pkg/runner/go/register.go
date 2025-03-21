package gorunner

import (
	"quitsh-cli/pkg/runner/config"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
)

// Register registers the runners in the factory.
func Register(
	lintSettings *config.LintSettings,
	factory factory.IFactory,
) (err error) {
	log.Trace("Register runner.", "id", GoLintRunnerID)
	e := factory.Register(
		GoLintRunnerID,
		runner.RunnerData{
			Creator: func(runnerConfig any) (runner.IRunner, error) {
				return NewGoLintRunner(runnerConfig, lintSettings)
			},
			RunnerConfigUnmarshal: UnmarshalLintConfig,
			DefaultToolchain:      "go-lint",
		})

	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey("lint", "go"), GoLintRunnerID)
	err = errors.Combine(err, e)

	return err
}
