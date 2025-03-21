package image

import (
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/registry"

	"github.com/hashicorp/go-version"
)

// NewImageRef construct a new image ref.
// This is the top-level function to generate image references.
// In the form:
// - `<base-name>-<registry-type>/<packageName>:<version>` for release image names.
// - `<base-name>-<registry-type>/<packageName>:<version>-<git-hash>` for non-release image names.
func NewImageRef(
	gitx git.Context,
	baseName string,
	packageName string,
	version *version.Version,
	registryType registry.RegistryType,
	isRelease bool,
) (ImageRef, error) {
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
