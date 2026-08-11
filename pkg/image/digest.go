package image

import (
	"github.com/containers/image/v5/docker/reference"
	"github.com/opencontainers/go-digest"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/skopeo"
)

// Sets the digest either from the `pkg.ImageFile` or from `pkg.ImageRef[0]`.
func (pkg *ImagePackage) SetDigest(
	ctx skopeo.Context,
	fromFile bool,
	transportProtocol string,
) error {
	// Get the digest.
	debug.Assert(len(pkg.ImageRefs) != 0, "package has no image ref")

	var ref string
	if fromFile {
		ref = "docker-archive://" + pkg.ImageFile
	} else {
		ref = transportProtocol + pkg.ImageRefs[0].Ref.String()
	}

	dig, err := ctx.Get("inspect", "--format", "{{.Digest}}", ref)
	if err != nil {
		return errors.AddContext(
			err,
			"could not get digest from ref '%s'",
			pkg.ImageRefs[0].Ref,
		)
	}

	namedRef, ok := pkg.ImageRefs[0].Ref.(ImageRefNamed)
	if !ok {
		return errors.New("image ref '%v' is not a named image ref", pkg.ImageRefs[0].Ref)
	}

	pkg.ImageDigest, err = digest.Parse(dig)
	if err != nil {
		return errors.AddContext(err, "could not parse image digest '%v'", dig)
	}
	pkg.ImageRefDigest.Ref, err = reference.WithDigest(namedRef, pkg.ImageDigest)
	if err != nil {
		return errors.AddContext(err, "could not compute image reference with digest")
	}

	return nil
}
