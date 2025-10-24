package step

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/tags"
)

type (
	Index int

	// Config is the the runner config.
	Config struct {
		Index Index `yaml:"-"`

		// Include this step given conditions.
		Include Include `yaml:"include"`

		// The runner with this name and stage name from the target will be used.
		Runner string `yaml:"runner"`
		// The runner given by this id which will used.
		RunnerID string `yaml:"runnerID"`

		// The toolchain to use for the runner can be overridden.
		Toolchain string `yaml:"toolchain"`

		// The (optional) raw runner config, before unmarshalling.
		ConfigRaw AuxConfigRaw `yaml:"config,omitempty"`

		// Additional stuff not parsed by quitsh, but for general purposes
		// such as anchors etc. Needed due to strict parsing.
		DotGeneral any `yaml:".general,omitempty"`
	}

	Include struct {
		// The tag expression associated with this step.
		// This is useful to include/exclude this step given certain tags.
		// This expression is parsed as a Go build line `//go:build <tagInclude>.
		// If the expression matches against given tags, then this step is included.
		// If there is no expression (default) its included.
		TagExpr tags.Expr `yaml:"tagExpr"`
	}

	AuxConfigRaw struct {
		Unmarshal func(any) error `yaml:"-"`
	}

	// AuxConfig is the unmarshalled additional config,
	// (not useful, only to make typing clearer).
	AuxConfig any

	// RunnerConfigUnmarshaller is the interface registered in the
	// runner factory to unmarshal configs for the runner.
	RunnerConfigUnmarshaller func(raw AuxConfigRaw) (AuxConfig, error)
)

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

func (s *AuxConfigRaw) UnmarshalYAML(unmarshal func(any) error) error {
	// Save the unmarshal function for later use.
	s.Unmarshal = unmarshal

	return nil
}
