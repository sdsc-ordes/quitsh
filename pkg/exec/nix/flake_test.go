//go:build test && (test_small || test_all)

package nix

import (
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlakeOutputs(t *testing.T) {
	_, rootDir, _ := git.NewCtxAtRoot(".")

	ps, err := GetFlakeOutputs(rootDir, "tools/nix", []string{"devShells", "packages"})
	require.NoError(t, err)
	assert.NotEmpty(t, ps)
}

func TestFlakePackages(t *testing.T) {
	_, rootDir, _ := git.NewCtxAtRoot(".")

	ps, err := GetFlakePackages(rootDir, "tools/nix")
	require.NoError(t, err)
	assert.NotEmpty(t, ps)
}

func TestFlakeShells(t *testing.T) {
	_, rootDir, _ := git.NewCtxAtRoot(".")

	ps, err := GetFlakeShells(rootDir, "tools/nix")
	require.NoError(t, err)
	assert.NotEmpty(t, ps)
}
