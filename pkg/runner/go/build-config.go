package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
)

type RunnerConfigBuild struct {
	VersionModule string `yaml:"versionModule" default:"pkg/build"`

	// Relative paths to sub modules to run `go test -C <compPath>/<path> <compPath>/<path>/...`
	// If specified add `.` to include the component as well.
	Submodules []string `yaml:"submodules" default:"[]"`

	// Additional build tags.
	BuildTags []string `yaml:"buildTags" default:"[]"`
}

func (c *RunnerConfigBuild) Validate() error {
	return common.Validator().Struct(c)
}

// UnmarshalBuildConfig is the unmarshaller for the BuildConfig.
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
