//go:build test

package echorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
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
	log.Trace("Register runner.", "id", EchoRunnerID)

	e := factory.Register(
		EchoRunnerID,
		runner.RunnerData{
			Creator: func(config step.AuxConfig) (runner.IRunner, error) {
				return NewEchoRunner(config, buildSettings)
			},
			RunnerConfigUnmarshal: UnmarshalEchoConfig,
			DefaultToolchain:      "go",
		})
	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey("build", "echo"), EchoRunnerID)
	err = errors.Combine(err, e)

	return err
}
