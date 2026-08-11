package coverage

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
	testSettings config.ITestSettings,
	factory factory.IFactory,
) (err error) {
	log.Trace("Register runner.", "id", CoverageUploadRunnerID)
	e := factory.Register(
		CoverageUploadRunnerID,
		runner.RunnerData{
			Creator: func(config step.AuxConfig) (runner.IRunner, error) {
				return NewCodecovRunner(config, testSettings)
			},
			RunnerConfigUnmarshal: UnmarshalCodecovConfig,
			DefaultToolchain:      "coverage-upload",
		})

	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey(stage.Test,
		"coverage-upload"), CoverageUploadRunnerID)
	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey(stage.Coverage,
		"coverage-upload"), CoverageUploadRunnerID)
	err = errors.Combine(err, e)

	return err
}
