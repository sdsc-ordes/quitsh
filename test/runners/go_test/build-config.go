//go:build test

package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
)

type GoBuildConfig struct {
	VersionModule string `yaml:"versionModule" default:"pkg/build"`

	// Additional build tags.
	BuildTags []string `yaml:"buildTags" default:"[]"`
}

func (c *GoBuildConfig) Validate() error {
	return common.Validator().Struct(c)
}

// UnmarshalBuildConfig is the unmarshaller for the [GoBuildConfig].
func UnmarshalBuildConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &GoBuildConfig{}
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
