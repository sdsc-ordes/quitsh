package fs

import (
	"os"
	"path"
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeDirs(t testing.TB, dir string) {
	type d struct {
		p   string
		dir bool
	}

	paths := []d{
		{path.Join(dir, "a/a-1/a-2"), true},
		{path.Join(dir, "a/a-3"), true},

		{path.Join(dir, "a/f.txt"), false},
		{path.Join(dir, "a/ignored"), false},
		{path.Join(dir, "a/a-1/g.txt"), false},
		{path.Join(dir, "a/a-1/a-2/h.txt"), false},

		{path.Join(dir, "b/b-1/b-2"), true},
		{path.Join(dir, "b/b-3"), true},

		{path.Join(dir, "b/f.txt"), false},
		{path.Join(dir, "b/ignored"), false},
		{path.Join(dir, "b/b-1/g.txt"), false},
		{path.Join(dir, "b/b-1/b-2/h.txt"), false},
	}

	makePaths := func(p d) {
		if p.dir {
			e := os.MkdirAll(p.p, DefaultPermissionsDir)
			require.NoError(t, e)
		} else {
			f, e := os.Create(p.p)
			require.NoError(t, e)
			_ = f.Close()
		}
	}

	for _, p := range paths {
		makePaths(p)
	}
}

func TestGlob(t *testing.T) {
	err := log.Setup("trace")
	require.NoError(t, err)

	dir := t.TempDir()
	makeDirs(t, dir)

	{
		files, traverseFiles, e := FindFiles(dir,
			WithPathFilterPatterns([]string{"**/*.txt"}, []string{"**/h.*"}, true))
		require.NoError(t, e)
		assert.Equal(t, int64(8), traverseFiles)

		for _, s := range []string{
			path.Join(dir, "a/f.txt"),
			path.Join(dir, "a/a-1/g.txt"),
			path.Join(dir, "b/f.txt"),
			path.Join(dir, "b/b-1/g.txt")} {
			assert.Contains(t, files, s)
		}
		assert.Len(t, files, 4)
	}

	{
		files, traverseFiles, e := FindFiles(dir,
			WithPathFilterPatterns(nil,
				[]string{"**/h.*", "**/g.txt", "**/ignore*"}, true))

		require.NoError(t, e)
		assert.Equal(t, int64(8), traverseFiles)

		for _, s := range []string{
			path.Join(dir, "a/f.txt"),
			path.Join(dir, "b/f.txt")} {
			assert.Contains(t, files, s)
		}
		assert.Len(t, files, 2)
	}

	{
		files, traverseFiles, e := FindFiles(dir,
			WithWalkDirFilterPatterns(nil, []string{"**/*-1"}, true),
			WithPathFilterPatterns([]string{"**/*.txt"}, nil, true))

		require.NoError(t, e)
		assert.Equal(t, int64(4), traverseFiles)

		for _, s := range []string{
			path.Join(dir, "a/f.txt"),
			path.Join(dir, "b/f.txt")} {
			assert.Contains(t, files, s)
		}
		assert.Len(t, files, 2)
	}

	{
		// And connection.
		files, traverseFiles, e := FindFiles(dir,
			WithPathFilterPatterns([]string{"**/f.*"}, nil, true),
			WithPathFilterPatterns([]string{"**/g.*"}, nil, true))

		require.NoError(t, e)
		assert.Equal(t, int64(8), traverseFiles)

		assert.Empty(t, files)
	}

	{
		// Or connection.
		files, traverseFiles, e := FindFiles(dir,
			WithPathFilterPatterns([]string{"**/f.*"}, nil, true),
			WithPathFilterPatterns([]string{"**/g.*"}, nil, false))

		require.NoError(t, e)
		assert.Equal(t, int64(8), traverseFiles)
		for _, s := range []string{
			path.Join(dir, "a/f.txt"),
			path.Join(dir, "a/a-1/g.txt"),
			path.Join(dir, "a/f.txt"),
			path.Join(dir, "b/b-1/g.txt")} {
			assert.Contains(t, files, s)
		}

		assert.Len(t, files, 4)
	}
}

func TestGlobDirs(t *testing.T) {
	err := log.Setup("trace")
	require.NoError(t, err)

	dir := t.TempDir()
	makeDirs(t, dir)
	{
		// Walk dirs only.
		paths, _, e := FindPaths(dir,
			WithReportOnlyDirs(true),
			// a-3 and b-3 are skipped.
			WithWalkDirFilterPatterns(nil, []string{"**/*-3"}, true),
			// only report a-2 and b-2
			WithPathFilterPatterns([]string{"**/*-2"}, nil, true))

		require.NoError(t, e)
		for _, s := range []string{
			path.Join(dir, "a/a-1/a-2"),
			path.Join(dir, "b/b-1/b-2")} {
			assert.Contains(t, paths, s)
		}

		assert.Len(t, paths, 2)
	}
}
