package config

import "github.com/sdsc-ordes/quitsh/pkg/common"

type ITestSettings interface {
	// The build type for the tests.
	BuildType() common.BuildType

	// Show the test log of the tests.
	ShowTestLog() bool

	// Additional arguments forwarded to the test tool.
	Args() []string
}
