package image

import (
	"fmt"
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
	domain string,
	basePathFmt string,
	packageName string,
	version *version.Version,
	registryType registry.Type,
	isRelease bool,
) (ImageRef, error) {
	if domain == "" || basePathFmt == "" {
		return nil, errors.New("domain or base path for image ref must not be empty")
	}

	if registryType == registry.RegistryTiltRegistry {
		domain = os.Getenv("EXPECTED_REGISTRY")
		if domain == "" {
			return nil, errors.New(
				"could not get expected image registry from " +
					"'tilt' env. variable 'EXPECTED_REGISTRY', is it defined?",
			)
		}
	}

	tag := version.String()
	if tag == "" {
		// Should not be the case:
		// See: https://github.com/hashicorp/go-version/issues/170
		return nil, errors.New("version is empty, cannot create image ref")
	}

	if !isRelease {
		commitSHA, err := gitx.Get("rev-parse", "--short=12", "HEAD")
		if err != nil {
			return nil, err
		}
		tag += "-" + commitSHA
	}

	//NOTE: We cannot use baseName/registryType, as somehow only
	//      one level of directory is allowed.
	return NewRef(
		path.Join(domain, fmt.Sprintf(basePathFmt, registryType.String()), packageName),
		tag,
		"",
	)
}
