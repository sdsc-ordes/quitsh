package fs

import (
	"os"
	"path"
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestGlob(t *testing.T) {
	err := log.Setup("trace")
	require.NoError(t, err)

	dir := t.TempDir()
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
			f.Close()
		}
	}

	for _, p := range paths {
		makePaths(p)
	}

	{
		files, traverseFiles, e := FindFiles(dir,
			WithGlobFilePatterns([]string{"**/*.txt"}, []string{"**/h.*"}, true))
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
			WithGlobFilePatterns(nil,
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
			WithGlobDirPatterns(nil, []string{"**/*-1"}, true),
			WithGlobFilePatterns([]string{"**/*.txt"}, nil, true))

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
			WithGlobFilePatterns([]string{"**/f.*"}, nil, true),
			WithGlobFilePatterns([]string{"**/g.*"}, nil, true))

		require.NoError(t, e)
		assert.Equal(t, int64(8), traverseFiles)

		assert.Empty(t, files)
	}

	{
		// Or connection.
		files, traverseFiles, e := FindFiles(dir,
			WithGlobFilePatterns([]string{"**/f.*"}, nil, true),
			WithGlobFilePatterns([]string{"**/g.*"}, nil, false))

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
