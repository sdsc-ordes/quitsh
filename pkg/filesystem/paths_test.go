package fs

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.True(t, Exists(dir))
	assert.False(t, Exists(path.Join(dir, "asdf")))

	e, err := ExistsE(dir)
	require.NoError(t, err)
	assert.True(t, e)

	e, err = ExistsE(path.Join(dir, "asdf"))
	require.Error(t, err)
	assert.False(t, e)
}

func TestPathExistsLinks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dir2 := t.TempDir()

	link := path.Join(dir2, "link")
	err := os.Symlink(dir, link)
	require.NoError(t, err)
	assert.True(t, Exists(link))

	// Remove the dir, symlink is dangling.
	err = os.RemoveAll(dir)
	require.NoError(t, err)
	assert.False(t, Exists(link))
}

func TestMakeAbs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	d := path.Join(dir, "a/b/c")
	err := os.MkdirAll(d, DefaultPermissionsDir)
	require.NoError(t, err)

	p, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(p) }()

	err = os.Chdir(d)
	require.NoError(t, err)

	assert.Equal(t, d, MakeAbsolute("."))
	assert.Equal(t, []string{d, path.Join(d, "a")}, MakeAllAbsolute(".", "./a"))
}

func TestMakeRelativeTo(t *testing.T) {
	t.Parallel()

	p, e := MakeRelativeTo("/a/b/c/d/e//../..", "/a/b/c/d/e")
	require.NoError(t, e)
	assert.Equal(t, "d/e", p)

	p, e = MakeRelativeTo("b/c", "b/c/d")
	require.NoError(t, e)
	assert.Equal(t, "d", p)

	p, e = MakeRelativeTo("b", "b/c/d")
	require.NoError(t, e)
	assert.Equal(t, "c/d", p)

	p, e = MakeRelativeTo("a/b/c", "b/c")
	require.NoError(t, e)
	assert.Equal(t, "../../../b/c", p)
}
