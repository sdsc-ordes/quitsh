package git

import (
	"os"
	"path"
	"testing"

	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeDirs(t testing.TB, dir string) {
	type d struct {
		p string
	}

	paths := []d{
		{path.Join(dir, "a.txt")},
		{path.Join(dir, "b.txt")},
		{path.Join(dir, "c.txt")},
		{path.Join(dir, "d.txt")},
		{path.Join(dir, "h.txt")},
	}

	makePaths := func(p d) {
		f, e := os.Create(p.p)
		require.NoError(t, e)
		_ = f.Close()
	}

	for _, p := range paths {
		makePaths(p)
	}

	e := os.Symlink(path.Join(dir, "a.txt"), path.Join(dir, "symlink.txt"))
	require.NoError(t, e)
}

func TestFind(t *testing.T) {
	err := log.Setup("trace")
	require.NoError(t, err)

	dir := t.TempDir()
	makeDirs(t, dir)
	gitx := NewCtx(dir, adjustGitCtx)

	err = gitx.Check("init")
	require.NoError(t, err)
	err = os.WriteFile(
		path.Join(dir, ".gitignore"), []byte("d.txt"),
		fs.DefaultPermissionsFile)
	require.NoError(t, err)

	gitTracked := []string{
		path.Join(dir, "a.txt"),
		path.Join(dir, "b.txt"),
		path.Join(dir, "symlink.txt"),
		path.Join(dir, ".gitignore"),
	}
	for _, f := range gitTracked {
		err = gitx.Check("add", f)
		require.NoError(t, err)
	}
	err = gitx.Check("commit", "-a", "-m", "init")
	require.NoError(t, err)

	// Modify a file.
	err = os.WriteFile(
		path.Join(dir, "a.txt"), []byte("bla"),
		fs.DefaultPermissionsFile)
	require.NoError(t, err)

	expected := []string{
		path.Join(dir, "a.txt"),
		path.Join(dir, "b.txt"),
		path.Join(dir, "c.txt"),
		path.Join(dir, "symlink.txt"),
	}

	{
		files, _, e := gitx.FilterFilesByPatterns(
			[]string{"**/*.txt"}, []string{"**/h.*"},
		)
		require.NoError(t, e)

		for _, s := range expected {
			assert.Contains(t, files, s)
		}
		assert.Len(t, files, len(expected))
	}

	{
		files, _, e := gitx.FilterPaths(
			fs.WithOnlyRegularFiles(true),
			fs.WithPathFilterPatterns([]string{"**/*.txt"}, []string{"**/h.*"}, true),
		)
		require.NoError(t, e)

		for _, s := range expected[0:3] {
			assert.Contains(t, files, s)
		}
		assert.Len(t, files, len(expected[0:3]))
	}
}
