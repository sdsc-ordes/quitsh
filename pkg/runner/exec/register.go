package execrunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
)

// Register registers the runner.
func Register(
	buildSettings config.IBuildSettings,
	factory factory.IFactory,
	registerKey bool,
) (err error) {
	log.Trace("Register runner.", "id", ExecRunnerID)
	e := factory.Register(
		ExecRunnerID,
		runner.RunnerData{
			Creator: func(config step.AuxConfig) (runner.IRunner, error) {
				return NewExecRunner(config, buildSettings)
			},
			RunnerConfigUnmarshal: UnmarshalRunnerConfig,
			DefaultToolchain:      "runner-exec",
		})
	err = errors.Combine(err, e)

	if registerKey {
		s := factory.Stages()
		for i := range s {
			e = factory.RegisterToKey(runner.NewRegisterKey(s[i].Stage, "exec"), ExecRunnerID)
			err = errors.Combine(err, e)
		}
	}

	return err
}
