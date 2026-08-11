package trivyrunner

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
	log.Trace("Register runner.", "id", TrivyLintRunnerID)
	e := factory.Register(
		TrivyLintRunnerID,
		runner.RunnerData{
			Creator: func(config step.AuxConfig) (runner.IRunner, error) {
				return NewTrivyLintRunner(config, lintSettings)
			},
			RunnerConfigUnmarshal: UnmarshalTrivyConfig,
			DefaultToolchain:      "lint-trivy",
		})

	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey(stage.Lint, "trivy"), TrivyLintRunnerID)
	err = errors.Combine(err, e)

	return err
}
