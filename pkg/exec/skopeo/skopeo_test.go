//go:build test_large || test_all

package skopeo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSkopeo(t *testing.T) {
	ctx := NewCtx()
	e := ctx.Check("--version")
	require.NoError(t, e)
}

func TestSkopeoWithTLS(t *testing.T) {
	ctx := NewCtx()
	inspCtx := ctx.InspectCtx()
	e := inspCtx.Check("--no-creds", "docker://alpine:latest")
	require.NoError(t, e)
}
