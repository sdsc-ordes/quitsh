package image

import (
	"os"
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/registry"

	"github.com/hashicorp/go-version"
)

// NewImageRef construct a new image ref.
// This is the top-level function to generate image references.
// In the form:
// - `<base-name>-<registry-type>/<packageName>:<version>` for release image names.
// - `<base-name>-<registry-type>/<packageName>:<version>-<git-hash>` for non-release image names.
// - If the `registryType` is `RegistryTempTilt` it will be taken from `EXPECTED_REF` env. variable.
func NewImageRef(
	gitx git.Context,
	baseName string,
	packageName string,
	version *version.Version,
	registryType registry.Type,
	isRelease bool,
) (ImageRef, error) {
	if registryType == registry.RegistryTempTilt {
		ref := os.Getenv("EXPECTED_REF")
		if ref == "" {
			return nil, errors.New(
				"could not get expected image ref from " +
					"'tilt' env. variable 'EXPECTED_REF', is it defined?",
			)
		}

		return NewRefFromString(ref)
	}

	tag := version.String()

	if !isRelease {
		commitSHA, err := gitx.Get("rev-parse", "--short=12", "HEAD")
		if err != nil {
			return nil, err
		}
		tag += "-" + commitSHA
	}

	//NOTE: We cannot use baseName/registryType, as somehow only
	//      one level of directory is allowed.
	return NewRef(path.Join(baseName+"-"+registryType.String(), packageName), tag, "")
}
