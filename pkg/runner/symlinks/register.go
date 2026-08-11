package symlinkrunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
)

// Register registers the runners in the factory.
func Register(
	lintSettings *config.LintSettings,
	factory factory.IFactory,
) (err error) {
	log.Trace("Register runner.", "id", SymlinkLintRunnerID)

	e := factory.Register(
		SymlinkLintRunnerID,
		runner.RunnerData{
			Creator: func(config step.AuxConfig) (runner.IRunner, error) {
				return NewSymlinkLintRunner(config, lintSettings)
			},
			RunnerConfigUnmarshal: UnmarshalLintConfig,
			DefaultToolchain:      "ci",
		})

	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey(stage.Lint, "symlink"), SymlinkLintRunnerID)
	err = errors.Combine(err, e)

	return err
}
