//go:build test

package echorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

type EchoConfig struct {
	Text string
}

func (c *EchoConfig) Validate() error {
	return validator.New().Struct(c)
}

// UnmarshalEchoConfig is the unmarshaller for the [EchoConfig].
func UnmarshalEchoConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &EchoConfig{}
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
