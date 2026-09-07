package image

import (
	"fmt"

	"github.com/opencontainers/go-digest"
)

type (
	ImagePackages []*ImagePackage

	ImagePackage struct {
		Component string `yaml:"component"`
		Version   string `yaml:"version"`

		// The image package name. (Same as basename of the image ref)
		Name      string `yaml:"name"`
		ImageType Type   `yaml:"imageType"`

		// For Nix build.
		NixPackage     string `yaml:"nixPackage"`
		NixInstallable string `yaml:"nixInstallable,omitempty"`

		// For normal Containerfile build.
		ContainerFile string `yaml:"containerFile,omitempty"`
		ImageFile     string `yaml:"src,omitempty"`

		// The image digest of this image.
		ImageDigest digest.Digest `yaml:"imageDigest"`

		// The image ref with full digest.
		ImageRefDigest ImageRefField `yaml:"imageRefDigest"`

		// Image references, the first one is the image which is built, all
		// other ones are upload references.
		ImageRefs []ImageRefField `yaml:"imageRefs"`
	}

	ImagePackageNameF = func(compName string, imageType Type) string
)

// ImagePackageNameDefault return the components image package name for
// the image type.
func ImagePackageNameDefault(compName string, imageType Type) string {
	return fmt.Sprintf("%s-%s", compName, imageType.String())
}
