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
		domain      string
		path        string
		pkg         string
		regType     registry.Type
		isRel       bool
		expectedRef string
	}

	tiltReg := "localhost:4000"
	t.Setenv("EXPECTED_REGISTRY", tiltReg)
	defer func() { _ = os.Unsetenv("EXPECTED_REGISTRY") }()

	tests := []D{
		{
			domain:  "domain.com",
			path:    "a/b/c-%s",
			pkg:     "mypkg",
			regType: registry.RegistryRelease,
			isRel:   true,
		},
		{
			domain:  "domain.com",
			path:    "a/b/c-%s",
			pkg:     "mypkg",
			regType: registry.RegistryRelease,
			isRel:   false,
		},
		{
			domain:  "domain.com",
			path:    "a/b/c-%s",
			pkg:     "mypkg",
			regType: registry.RegistryTemp,
			isRel:   true,
		},
		{
			domain:  "domain.com",
			path:    "a/b/c-%s",
			pkg:     "mypkg",
			regType: registry.RegistryTemp,
			isRel:   false,
		},
		{
			domain:  "domain.com",
			path:    "a/b/c-%s",
			pkg:     "mypkg",
			regType: registry.RegistryTiltRegistry,
			isRel:   true,
		},
		{
			domain:  "domain.com",
			path:    "a/b/c-%s",
			pkg:     "mypkg",
			regType: registry.RegistryTiltRegistry,
			isRel:   false,
		},
	}

	for _, te := range tests {
		tag := v.String()
		if !te.isRel {
			tag += "-" + commitSHA
		}
		if te.regType == registry.RegistryTiltRegistry {
			te.domain = tiltReg
		}
		te.expectedRef = fmt.Sprintf(
			"%s/%s/%s:%s",
			te.domain,
			fmt.Sprintf(te.path, te.regType.String()),
			te.pkg,
			tag,
		)

		ref, e := NewImageRef( //nolint:govet //intentional
			te.domain,
			te.path,
			te.pkg,
			v,
			te.regType,
			commitSHA,
			te.isRel,
		)
		require.NoError(t, e)

		assert.Equal(
			t,
			te.expectedRef,
			ref.String(),
		)
	}
}
