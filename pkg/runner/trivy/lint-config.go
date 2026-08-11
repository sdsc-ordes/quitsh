package trivyrunner

import (
	"github.com/creasty/defaults"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
)

type (
	RunnerConfigLint struct {
		ConfigFile string `yaml:"configFile" default:"tools/configs/trivy/trivy.yaml"`
	}
)

func UnmarshalTrivyConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &RunnerConfigLint{} //nolint:exhaustruct // Intended
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
