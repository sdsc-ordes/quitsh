package fs

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	sub := path.Join(dir, "test1")
	e := os.MkdirAll(sub, DefaultPermissionsDir)
	require.NoError(t, e)

	dir2 := path.Join(t.TempDir(), "test")
	subDest := path.Join(dir2, "test1")

	e = CopyFileOrDir(dir, dir2, false)
	require.NoError(t, e)
	require.DirExists(t, subDest)

	e = CopyFileOrDir(dir, dir2, false)
	require.ErrorIs(t, e, os.ErrExist)
	require.DirExists(t, subDest)

	// Overwrite with some stuff.
	sub2 := path.Join(dir, "test2")
	subDest2 := path.Join(dir2, "test2")
	e = os.MkdirAll(sub2, DefaultPermissionsDir)
	require.NoError(t, e)
	e = CopyFileOrDir(dir, dir2, true)
	require.NoError(t, e)
	require.DirExists(t, subDest)
	require.DirExists(t, subDest2)
}
