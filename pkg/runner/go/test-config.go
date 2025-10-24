package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

type RunnerTestConfig struct {
	// Additional build tags.
	BuildTags []string `yaml:"buildTags" default:"[]"`

	// Additional arguments forwarded to the test tool (`go test`).
	Args []string `yaml:"args"`

	// Additional arguments forwarded to the test executable (`go test ... -args ...`).
	TestArgs []string `yaml:"testArgs"`
}

func (c *RunnerTestConfig) Validate() error {
	return validator.New().Struct(c)
}

// The unmarshaller for the BuildConfig.
func UnmarshalTestConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &RunnerTestConfig{}
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
