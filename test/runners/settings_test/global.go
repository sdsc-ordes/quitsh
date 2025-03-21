//go:build test

package settings

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
)

type BuildSettings struct {
	// The build type.
	BuildType common.BuildType `yaml:"buildType"`
}

// NewBuildSettings constructs a new build setting.
func NewBuildSettings(
	buildType common.BuildType,
) BuildSettings {
	return BuildSettings{
		BuildType: buildType,
	}
}
