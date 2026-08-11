package gorunner

import (
	"github.com/creasty/defaults"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
)

type (
	LintConfig struct {
		CheckBuildConstraints CheckBuildConstraints `yaml:"checkBuildConstraints"`
		GolangCILint          GolangCILint          `yaml:"golangCILint"`
	}

	CheckBuildConstraints struct {
		Enable bool                  `yaml:"enable"`
		Rules  []BuiltConstraintRule `yaml:"rules"`
	}

	GolangCILint struct {
		Config string   `yaml:"config" default:"tools/configs/golangci-lint/golangci.yaml"`
		Args   []string `yaml:"args"`
	}

	BuiltConstraintRule struct {
		// Which files to include by glob pattern.
		IncludePatterns []string `yaml:"includePatterns"`

		// The build constraint must match the following
		// string exactly.
		Constraints []string `yaml:"constraints"`
	}
)

func UnmarshalLintConfig(raw step.AuxConfigRaw) (step.AuxConfig, error) {
	config := &LintConfig{} //nolint: exhaustruct // intended
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
