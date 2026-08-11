package gitdiffrunner

import (
	"github.com/creasty/defaults"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
)

type (
	RunnerDiffConfig struct {
		// If set, only these paths are used for diffing.
		// Relative paths to root.
		RootPaths []string `yaml:"rootPaths"`
		// Paths relative to component.
		Paths []string `yaml:"paths"`
	}
)

// UnmarshalDiffConfig unmarshals for the [RunnerDiffConfig].
func UnmarshalDiffConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &RunnerDiffConfig{} //nolint:exhaustruct // Intended
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

	return config, nil
}
