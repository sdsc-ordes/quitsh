package cmdrunnner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

type RunnerConfig struct {
	Name string `yaml:"name" default:"unnamed"`

	// Either run a script piped to command `cmd`, which could be `bash -eu`.
	Script string `yaml:"script"`

	// ... or run a command directly.
	Cmd []string `yaml:"cmd"`

	// Env sets additional environment variables.
	Env []string `yaml:"env"`
}

func (c *RunnerConfig) Validate() error {
	return validator.New().Struct(c)
}

// UnmarshalRunnerConfig unmarshals [RunnerConfig].
func UnmarshalRunnerConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &RunnerConfig{}
	err := defaults.Set(config)
	if err != nil {
		return nil, err
	}

	// Deserialize if we have something.
	if raw.Unmarshal != nil {
		err = raw.Unmarshal(config)
		if err != nil {
			return nil, err
		}
	}

	err = config.Validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}
