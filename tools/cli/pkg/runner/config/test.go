package config

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

type TestSettings struct {
	// The build type for the tests.
	BuildType common.BuildType `yaml:"buildType"`

	// Show the test log of the tests.
	ShowTestLog bool `yaml:"showTestLog"`

	// Additional arguments forwarded to the test tool.
	Args []string `yaml:"args"`
}

// NewBuildSettings constructs a new build setting.
func NewTestSettings(
	buildType common.BuildType,
	showTestLog bool,
	args []string,
) TestSettings {
	return TestSettings{
		BuildType:   buildType,
		Args:        args,
		ShowTestLog: showTestLog,
	}
}

type wrapITestSettings struct {
	// NOTE: We cannot make `Build()` function and have a `BuildType` type
	// (needs to public due to YAML)
	ref *TestSettings
}

func (c *wrapITestSettings) BuildType() common.BuildType {
	return c.ref.BuildType
}
func (c *wrapITestSettings) ShowTestLog() bool {
	return c.ref.ShowTestLog
}
func (c *wrapITestSettings) Args() []string {
	return c.ref.Args
}

// WrapToITestSettings returns a interface for the quitsh runners.
func (b *TestSettings) WrapToITestSettings() config.ITestSettings {
	return &wrapITestSettings{ref: b}
}
