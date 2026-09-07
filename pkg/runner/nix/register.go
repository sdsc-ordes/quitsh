package nixrunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/image"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
)

type (
	Option func(*opts)

	opts struct {
		installable ImageInstallable
		packageName image.ImagePackageNameF
		stage       stage.Stage
	}
)

func (c *opts) Apply(options ...Option) {
	for _, f := range options {
		f(c)
	}

	if c.installable == nil {
		c.installable = ImageInstallableDefault
	}

	if c.packageName == nil {
		c.packageName = image.ImagePackageNameDefault
	}

	if c.stage == "" {
		c.stage = stage.Image
	}
}

func WithImageInstallable(f ImageInstallable) Option {
	return func(o *opts) {
		o.installable = f
	}
}

func WithStage(s stage.Stage) Option {
	return func(o *opts) {
		o.stage = s
	}
}

func WithImagePackageName(f image.ImagePackageNameF) Option {
	return func(o *opts) {
		o.packageName = f
	}
}

// Register registers the runners in the factory.
func Register(
	imageSettings *config.ImageSettings,
	nixSettings *config.NixSettings,
	factory factory.IFactory,
	options ...Option,
) (err error) {
	log.Trace("Register runner.", "id", NixImageRunnerID)

	var o opts
	o.Apply(options...)

	e := factory.Register(
		NixImageRunnerID,
		runner.RunnerData{
			Creator: func(config step.AuxConfig) (runner.IRunner, error) {
				return NewNixImageRunner(
					nixSettings.FlakeDirRel,
					config,
					imageSettings,
					nixSettings,
					&o,
				)
			},
			RunnerConfigUnmarshal: UnmarshalImageConfig,
			DefaultToolchain:      "image-nix",
		})
	err = errors.Combine(err, e)
	e = factory.RegisterToKey(runner.NewRegisterKey(o.stage, "nix"), NixImageRunnerID)
	err = errors.Combine(err, e)

	return
}
