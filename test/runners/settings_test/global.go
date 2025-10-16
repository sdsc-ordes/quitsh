//go:build test

package settings

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

type (
	BuildSettings struct {
		// The build type.
		BuildType common.BuildType `yaml:"buildType"`
		// The environment type.
		EnvironmentType common.EnvironmentType `yaml:"environmentType"`

		// If coverage information should be built into.
		Coverage bool `yaml:"coverage"`

		// Additional arguments handed to the build tool.
		Args []string `yaml:"args"`
	}

	wrapIBuildSettings struct {
		// NOTE: We cannot make `Build()` function and have a `BuildType` type
		// (needs to public due to YAML)
		ref *BuildSettings
	}
)

// NewBuildSettings constructs a new build setting.
func NewBuildSettings(
	buildType common.BuildType,
) BuildSettings {
	return BuildSettings{
		BuildType: buildType,
	}
}

func (c *wrapIBuildSettings) BuildType() common.BuildType {
	return c.ref.BuildType
}
func (c *wrapIBuildSettings) EnvironmentType() common.EnvironmentType {
	return c.ref.EnvironmentType
}
func (c *wrapIBuildSettings) Coverage() bool {
	return c.ref.Coverage
}
func (c *wrapIBuildSettings) Args() []string {
	return c.ref.Args
}

// WrapToIBuildSettings returns a interface for the quitsh runners.
func (b *BuildSettings) WrapToIBuildSettings() config.IBuildSettings {
	return &wrapIBuildSettings{ref: b}
}
