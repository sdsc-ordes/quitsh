//go:build test_large || test_all

package processcompose

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessComposeDevenv(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	logger := log.NewLogger("test")
	err := fs.CopyFileOrDir("./test/flake.nix", path.Join(d, "flake.nix"), true)
	require.NoError(t, err)
	err = fs.CopyFileOrDir("./test/flake.lock", path.Join(d, "flake.lock"), true)
	require.NoError(t, err)

	socketPathFile := path.Join("./test", ".pc-socket-path")

	pcCtx, err := Start(logger, d, d,
		"mynamespace.shells.test",
		WithSocketPathFile(socketPathFile),
	)
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

	fulfilled, err := pcCtx.WaitTill(
		ctx,
		logger,
		ProcessCond{Name: "httpbin", State: ProcessRunning},
		ProcessCond{Name: "keycloak", State: ProcessReady},

		// FIXME: Set that to completed once https://github.com/cachix/devenv/issues/2879
		// is fixed.
		ProcessCond{Name: "completed", State: ProcessRunning},
	)
	require.NoError(t, err)
	assert.True(t, fulfilled)
	assert.FileExists(t, socketPathFile)
}

func TestProcessComposeTimeout(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	logger := log.NewLogger("test")
	err := fs.CopyFileOrDir("./test/flake.nix", path.Join(d, "flake.nix"), true)
	require.NoError(t, err)
	err = fs.CopyFileOrDir("./test/flake.lock", path.Join(d, "flake.lock"), true)
	require.NoError(t, err)

	pcCtx, err := Start(logger,
		d, d, "mynamespace.shells.test",
		WithMustBeStarted(false))
	require.NoError(t, err)
	defer func() {
		log, e := os.ReadFile(pcCtx.LogFile())
		require.NoError(t, e)
		t.Log("Process Compose Log:\n", string(log))

		e = pcCtx.Stop()
		require.NoError(t, e)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	fulfilled, err := pcCtx.WaitTill(
		ctx,
		logger,
		ProcessCond{Name: "not-existing", State: ProcessRunning},
	)
	require.NoError(t, err)
	assert.False(t, fulfilled)
}
