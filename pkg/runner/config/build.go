package config

import "github.com/sdsc-ordes/quitsh/pkg/common"

type IBuildSettings interface {
	// The build type.
	BuildType() common.BuildType
	// The environment type.
	EnvironmentType() common.EnvironmentType

	// If coverage information should be built into.
	Coverage() bool

	// Additional arguments handed to the build tool.
	Args() []string
}
