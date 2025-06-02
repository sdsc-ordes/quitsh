package processcompose

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessComposeDevenv(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	err := fs.CopyFileOrDir("./test/flake.nix", path.Join(d, "flake.nix"), true)
	require.NoError(t, err)
	err = fs.CopyFileOrDir("./test/flake.lock", path.Join(d, "flake.lock"), true)
	require.NoError(t, err)

	pcCtx, err := Start(d, d, "mynamespace.shells.test", false)
	require.NoError(t, err)
	defer func() {
		log, e := os.ReadFile(pcCtx.LogFile())
		require.NoError(t, e)
		t.Log("Process Compose Log:\n", string(log))

		e = pcCtx.Stop()
		require.NoError(t, e)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	running, err := pcCtx.WaitTillRunning(ctx, "httpbin")
	require.NoError(t, err)
	assert.True(t, running)
}
