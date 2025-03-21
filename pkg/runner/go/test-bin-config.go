package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component/step"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

type RunnerConfigTestBin struct {
	VersionModule string `yaml:"versionModule" default:"pkg/build"`

	// Additional arguments for building the executable.
	// The Go module path (e.g. `a/b/c`) where the
	// executable is located. You can build multiple ones
	// if the path is not too specific.
	BuildPkg string `yaml:"buildPkg"`

	// Additional build tags for building the executable.
	BuildTags []string `yaml:"buildTags"`

	// Additional arguments for running the tests.
	// The Go module path (e.g. `a/b/c`) where the
	// tests are located to test the binary.
	// By default builds everything in component root.
	TestPkg string `yaml:"testPkg"`
	// This enables selecting the tests to build with `go test -tags=XXX`
	TestTags []string `yaml:"testTags"`
}

func (c *RunnerConfigTestBin) Validate() error {
	return validator.New().Struct(c)
}

// The unmarshaller for the TestBinConfig.
func UnmarshalTestBinConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &RunnerConfigTestBin{}
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
