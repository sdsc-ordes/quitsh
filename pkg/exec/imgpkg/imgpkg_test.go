package imgpkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImgpkg(t *testing.T) {
	ctx, e := NewCtx()
	require.NoError(t, e)
	e = ctx.Check("--version")
	require.NoError(t, e)
}
