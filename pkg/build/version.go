package build

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

// This string is set in the Go runner with
// `-ldflags -X ".../pkg/build.buildVersion=..."`.
// but here it needs to be correct to make other CLI users of
// quitsh have the correct version.
var buildVersion = "0.28.1" //nolint:gochecknoglobals // Allowed for version.

func GetBuildVersion() *version.Version {
	ver, err := version.NewVersion(buildVersion)
	if err != nil {
		panic(fmt.Sprintf("Build version '%v' is invalid.", ver))
	}

	return ver
}
