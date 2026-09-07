package containerfilerunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

type RunnerConfigContainerfile struct {
	Containerfile string `yaml:"containerfile"`
}

func (c *RunnerConfigContainerfile) Validate() error {
	return validator.New().Struct(c)
}

func UnmarshalImageConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &RunnerConfigContainerfile{} //nolint:exhaustruct // Intended.
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
