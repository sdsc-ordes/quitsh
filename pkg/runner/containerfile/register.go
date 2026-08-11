package containerfilerunner

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
		packageName image.ImagePackageNameF
		stage       stage.Stage
	}
)

func (c *opts) Apply(options ...Option) {
	for _, f := range options {
		f(c)
	}

	if c.packageName == nil {
		c.packageName = image.ImagePackageNameDefault
	}
}

func WithImagePackageName(f image.ImagePackageNameF) Option {
	return func(o *opts) {
		o.packageName = f
	}
}

func WithStage(s stage.Stage) Option {
	return func(o *opts) {
		o.stage = s
	}
}

// Register registers the runners in the factory.
func Register(
	imageSettings *config.ImageSettings,
	factory factory.IFactory,
	options ...Option,
) (err error) {
	log.Trace("Register runner.", "id", ContainerfileRunnerID)

	var o opts
	o.Apply(options...)

	e := factory.Register(
		ContainerfileRunnerID,
		runner.RunnerData{
			Creator: func(config step.AuxConfig) (runner.IRunner, error) {
				return NewContainerfileBuildRunner(config, imageSettings, &o)
			},
			RunnerConfigUnmarshal: UnmarshalImageConfig,
			DefaultToolchain:      "image-containerfile",
		})
	err = errors.Combine(err, e)
	e = factory.RegisterToKey(
		runner.NewRegisterKey(o.stage, "containerfile"),
		ContainerfileRunnerID,
	)
	err = errors.Combine(err, e)

	return
}
