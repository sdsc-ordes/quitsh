package query

import (
	"os"
	"path"
	"testing"
	"text/template"

	"github.com/sdsc-ordes/quitsh/pkg/component"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFiles(t *testing.T) (string, []string, []string) {
	compTmpl := `
name: "{{ .Name }}"
version: "1.0.0"
language: go
`

	writeConfigFile := func(name, dir string) {
		e := os.MkdirAll(dir, fs.DefaultPermissionsDir)
		require.NoError(t, e)

		s, e := template.New("test").Parse(compTmpl)
		require.NoError(t, e)

		f, e := os.OpenFile(
			path.Join(dir, component.ConfigFilename),
			os.O_TRUNC|os.O_WRONLY|os.O_CREATE,
			fs.DefaultPermissionsFile,
		)
		require.NoError(t, e, "opening config file")
		defer f.Close()

		type D struct {
			Name string
		}

		e = s.Execute(f, D{Name: name})
		require.NoError(t, e, "writing config")
	}

	dir := t.TempDir()
	dirs := []string{
		path.Join(dir, "a"),
		path.Join(dir, "b"),
		path.Join(dir, "a/sub-1"),
		path.Join(dir, "c/sub-2"),
		path.Join(dir, "a/sub-1/e"),
	}

	names := []string{
		"a",
		"b",
		"a-sub-1",
		"c-sub-2",
		"a-sub-1-e",
	}

	for i, d := range dirs {
		writeConfigFile(names[i], d)
	}

	return dir, dirs, names
}

func findNames(t *testing.T, comps []*component.Component, names []string) {
	for _, c := range comps {
		n := c.Name()
		assert.Contains(
			t,
			names,
			n,
			"components found should contain '%v'",
			n,
		)
	}
}

func TestComponentFindInside(t *testing.T) {
	t.Parallel()
	err := log.Setup("debug")
	require.NoError(t, err)
	_, dirs, names := setupFiles(t)
	cG := component.NewComponentCreator("", nil)

	for i, d := range dirs {
		log.Info("Find", "d", d)
		comp, e := FindInside(d, cG)
		require.NoError(t, e, "should find it")
		assert.Equal(t, names[i], comp.Name())
	}

	for i, d := range dirs {
		dd := path.Join(d, "1", "2", "3")
		log.Info("Find", "d", dd)
		_ = os.MkdirAll(dd, fs.DefaultPermissionsDir)
		comp, e := FindInside(dd, cG)
		require.NoError(t, e, "should find it")
		assert.Equal(t, names[i], comp.Name())
	}
}

func TestComponentFindByPattern(t *testing.T) {
	t.Parallel()
	err := log.Setup("debug")
	require.NoError(t, err)
	dir, _, names := setupFiles(t)
	cG := component.NewComponentCreator("", nil)

	log.Info("Find all 5.")
	comps, _, err := FindByPatterns(dir, []string{"*"}, 5, cG, nil)
	require.NoError(t, err, "should find 5")
	assert.Len(t, comps, 5, "should return 5 results")
	findNames(t, comps, names)

	log.Info("Find all and fail.")
	_, _, err = FindByPatterns(dir, []string{"*"}, 6, cG, nil)
	require.ErrorContains(t, err, "min. count '6' components not found in")

	log.Info("Find all with excludes.")
	comps, _, err = FindByPatterns(dir, []string{"*", "!*-sub-*"}, 2, cG, nil)
	require.NoError(t, err, "should find it")
	assert.Len(t, comps, 2, "should return 2 results")
	findNames(t, comps, names[0:2])

	log.Info("Find one.")
	comps, _, err = FindByPatterns(dir, []string{"a"}, 0, cG, nil)
	require.NoError(t, err, "should find it")
	assert.Len(t, comps, 1, "should return 1 results")
	findNames(t, comps, names[0:1])

	log.Info("Find sub.")
	comps, _, err = FindByPatterns(dir, []string{"*-sub-*"}, 0, cG, nil)
	require.NoError(t, err, "should find it")
	assert.Len(t, comps, 3, "should return 3 results")
	findNames(t, comps, names[2:])
}

func TestComponentFindByPatternZero(t *testing.T) {
	t.Parallel()

	e := log.Setup("debug")
	require.NoError(t, e)
	dir, dirs, _ := setupFiles(t)
	cG := component.NewComponentCreator("", nil)

	for _, d := range dirs {
		comps, _, err := FindByPatterns(dir,
			[]string{"*"}, 1, cG,
			WithComponentDirSingle(d, true))
		require.NoError(t, err, "should find it")
		assert.Len(t, comps, 1, "should return 1 results '%s'", d)
	}

	_, _, e = FindByPatterns(dir,
		[]string{"*"}, 1, cG,
		WithComponentDirSingle("non-existing", true))
	require.Error(t, e, "min. count '1' components not found in")
}
