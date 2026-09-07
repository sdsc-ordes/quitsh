package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
)

type RunnerTestConfig struct {
	// Relative paths to sub modules to run `go test -C <compPath>/<path> <compPath>/<path>/...`
	// If specified add `.` to include the component as well.
	Submodules []string `yaml:"submodules" default:"[]"`

	// GOWORK settings, default is disabled.
	GoWork string `yaml:"goWork" default:"off"`
	// GOTOOLCHAIN settings, default is local.
	GoToolchain string `yaml:"goToolchain" default:"local"`

	// Additional build tags.
	BuildTags []string `yaml:"buildTags" default:"[]"`

	// Additional arguments forwarded to the test tool (`go test`).
	Args []string `yaml:"args"`

	// Additional arguments forwarded to the test executable (`go test ... -args ...`).
	TestArgs []string `yaml:"testArgs"`
}

func (c *RunnerTestConfig) Validate() error {
	return common.Validator().Struct(c)
}

// UnmarshalTestConfig is the unmarshaller for the BuildConfig.
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
