package step

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"
)

type Index int

// Config is the the runner config.
type Config struct {
	Index Index `yaml:"-"`

	// The runner with this name and stage name from the target will be used.
	Runner string `yaml:"runner"`
	// The runner given by this id which will used.
	RunnerID string `yaml:"runnerID"`

	// The toolchain to use for the runner can be overridden.
	Toolchain string `yaml:"toolchain"`

	// The (optional) raw runner config, before unmarshalling.
	ConfigRaw AuxConfigRaw `yaml:"config,omitempty"`
}

func (c *Config) Init(idx Index) (err error) {
	c.Index = idx

	if c.Runner != "" && c.RunnerID != "" || c.Runner == "" && c.RunnerID == "" {
		err = errors.Combine(
			err,
			errors.New(
				"you must either define 'runner' or 'runnerID' but not both",
			),
		)
	}

	return
}

type AuxConfigRaw struct {
	Unmarshal func(any) error `yaml:"-"`
}

// The unmarshalled additional config,
// (not useful, only to make typing clearer).
type AuxConfig interface{}

// RunnerConfigUnmarshaller is the interface registered in the
// runner factory to unmarshal configs for the runner.
type RunnerConfigUnmarshaller func(raw AuxConfigRaw) (AuxConfig, error)

func (s *AuxConfigRaw) UnmarshalYAML(unmarshal func(any) error) error {
	// Save the unmarshal function for later use.
	s.Unmarshal = unmarshal

	return nil
}
