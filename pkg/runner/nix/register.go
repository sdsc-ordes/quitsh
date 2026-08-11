package nixrunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
)

// Register registers the runners in the factory.
func Register(
	imageSettings *config.ImageSettings,
	nixSettings *config.NixSettings,
	factory factory.IFactory,
) (err error) {
	log.Trace("Register runner.", "id", NixImageRunnerID)

	e := factory.Register(
		NixImageRunnerID,
		runner.RunnerData{
			Creator: func(config step.AuxConfig) (runner.IRunner, error) {
				return NewNixImageRunner(
					nixSettings.FlakeDirRel,
					config,
					imageSettings,
					nixSettings,
				)
			},
			RunnerConfigUnmarshal: UnmarshalImageConfig,
			DefaultToolchain:      "image-nix",
		})
	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey("image", "nix"), NixImageRunnerID)
	err = errors.Combine(err, e)

	return
}
