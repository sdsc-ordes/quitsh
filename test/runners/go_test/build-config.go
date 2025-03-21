//go:build test

package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

type RunnerConfigBuild struct {
	VersionModule string `yaml:"versionModule" default:"pkg/build"`

	// Additional build tags.
	BuildTags []string `yaml:"buildTags" default:"[]"`
}

func (c *RunnerConfigBuild) Validate() error {
	return validator.New().Struct(c)
}

// The unmarshaller for the BuildConfig.
func UnmarshalBuildConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &RunnerConfigBuild{}
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
