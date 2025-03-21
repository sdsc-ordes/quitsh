package runner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
)

// The unique id the runner is registered on.
type RegisterID = string

// The key the runner is registered additionally per stage.
type RegisterKey struct {
	stage stage.Stage
	name  string
}

func NewRegisterKey(stage stage.Stage, runnerName string) RegisterKey {
	return RegisterKey{stage, runnerName}
}

func (r *RegisterKey) Stage() stage.Stage {
	return r.stage
}

func (r *RegisterKey) Name() string {
	return r.name
}

type RunnerData struct {
	// The Creator function for the runner.
	// The `runnerConfig` is the unmarshalled config from the
	// `component.Config`.
	Creator func(runnerConfig any) (IRunner, error)

	// The config unmarshaller for additional `config:` section
	// (can be nil if no config should be parsed).
	// The result will be passed to the creator.
	RunnerConfigUnmarshal step.RunnerConfigUnmarshaller

	// The default toolchain to use if not specified in config.
	DefaultToolchain string
}

type RegisterFunc = func(
	key RegisterKey,
	data ...RunnerData,
)
