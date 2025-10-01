package version

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-version"
)

// Bump the version `v` with `level` one of [`patch`, `minor`, `major`].
func Bump(
	v *version.Version,
	level string,
	prerelease string,
	buildMeta string) (*version.Version, error) {
	segments := v.Segments()
	for len(segments) < 3 {
		segments = append(segments, 0)
	}

	switch level {
	case "major":
		segments[0]++
		segments[1], segments[2] = 0, 0
	case "minor":
		segments[1]++
		segments[2] = 0
	case "patch":
		segments[2]++
	default:
		return nil, errors.New("version bump level is incorrect")
	}

	newV := fmt.Sprintf("%d.%d.%d", segments[0], segments[1], segments[2])
	if prerelease != "" {
		newV += "-" + prerelease
	}
	if buildMeta != "" {
		newV += "+" + buildMeta
	}

	newVer, err := version.NewVersion(newV)
	if err != nil {
		return nil, err
	}

	return newVer, nil
}
