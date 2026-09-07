package symlinkrunner

import (
	"github.com/creasty/defaults"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
)

type (
	SymlinkConfig struct{}
)

func UnmarshalLintConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &SymlinkConfig{}
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
