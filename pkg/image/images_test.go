package image

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageRef(t *testing.T) {
	gitx := git.NewCtx("")

	v, e := version.NewVersion("1.0.1")
	require.NoError(t, e)

	commitSHA, e := gitx.Get("rev-parse", "--short=12", "HEAD")
	require.NoError(t, e)

	type D struct {
		b     string
		p     string
		r     registry.Type
		isRel bool
	}

	t.Setenv("EXPECTED_REF", "tilt-chosen:tag")
	defer os.Unsetenv("EXPECTED_REF")

	tests := []D{
		{b: "a/b/c", p: "mypkg", r: registry.RegistryRelease, isRel: true},
		{b: "a/b/c", p: "mypkg", r: registry.RegistryRelease, isRel: false},
		{b: "a/b/c", p: "mypkg", r: registry.RegistryTemp, isRel: true},
		{b: "a/b/c", p: "mypkg", r: registry.RegistryTemp, isRel: false},
		{b: "a/b/c", p: "mypkg", r: registry.RegistryTempTilt, isRel: true},
		{b: "a/b/c", p: "mypkg", r: registry.RegistryTempTilt, isRel: false},
	}

	for _, te := range tests {
		ref, e := NewImageRef(gitx, te.b, te.p, v, te.r, te.isRel) //nolint:govet //intentional
		require.NoError(t, e)

		if te.r == registry.RegistryTempTilt {
			assert.Equal(t, "tilt-chosen:tag", ref.String())

			continue
		}

		tag := v.String()
		if !te.isRel {
			tag += "-" + commitSHA
		}

		assert.Equal(t, fmt.Sprintf("%s-%s/%s:%s", te.b, te.r.String(), te.p, tag), ref.String())
	}
}
