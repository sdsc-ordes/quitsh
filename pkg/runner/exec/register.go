package cmdrunnner

import (
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
	log.Trace("Register runner.", "id", CmdRunnerID)
	e := factory.Register(
		CmdRunnerID,
		runner.RunnerData{
			Creator: func(runnerConfig any) (runner.IRunner, error) {
				return NewCmdRunner(runnerConfig, buildSettings)
			},
			RunnerConfigUnmarshal: UnmarshalRunnerConfig,
			DefaultToolchain:      "cmd-runner",
		})
	err = errors.Combine(err, e)

	if registerKey {
		s := factory.Stages()
		for i := range s {
			e = factory.RegisterToKey(runner.NewRegisterKey(s[i].Stage, "cmd"), CmdRunnerID)
			err = errors.Combine(err, e)
		}
	}

	return err
}
