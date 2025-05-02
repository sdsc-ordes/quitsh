package component

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/config"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentsConfig(t *testing.T) {
	expected := `
name: comp1
version: "1.0.0"
language: go
targets:
  build:
    steps:
      - runner: build-go
  lint:
    steps:
      - runner: lint-go
  deploy:
    steps:
      - runnerID: asdf
  test:
    steps:
      - runner: build-go
`
	d := t.TempDir()
	rfile := path.Join(d, "test.yaml")

	e := os.WriteFile(rfile, []byte(expected), fs.DefaultPermissionsFile)
	require.NoError(t, e)

	rr, e := os.OpenFile(rfile, os.O_RDONLY, fs.DefaultPermissionsFile)
	require.NoError(t, e)
	c, e := config.LoadFromReader[Config](rr)
	rr.Close()

	require.NoError(t, e)
	assert.Equal(t, "comp1", c.Name)
	assert.Equal(t, "build-go", c.Targets["build"].Steps[0].Runner)
	assert.Equal(t, "lint-go", c.Targets["lint"].Steps[0].Runner)
	assert.Equal(t, "", c.Targets["deploy"].Steps[0].Runner)
	assert.Equal(t, "build-go", c.Targets["test"].Steps[0].Runner)

	wfile := path.Join(d, "test2.yaml")
	require.NoError(t, e)
	e = config.SaveToFile(wfile, &c)
	require.NoError(t, e)
}

func TestComponentsConfigFail(t *testing.T) {
	file := `
name: comp1
version: "1.a"
language: go
targets:
  build:
    steps:
    - runner: go
`
	f := strings.NewReader(file)
	_, e := config.LoadFromReader[Config](f)

	assert.ErrorContains(t, e, "Malformed version:", e)
}

func TestComponentsConfigFail2(t *testing.T) {
	file := `
name: ""
version: "1.0.0-rc1+build.1234"
targets:
  build:
    steps:
    - runner: go
`
	f := strings.NewReader(file)
	_, e := config.LoadFromReader[Config](f)

	require.ErrorContains(t, e, "Field validation for 'Name' failed", e)
	require.ErrorContains(t, e, "Field validation for 'Language' failed", e)
}

func TestComponentsConfigRunnerConfig(t *testing.T) {
	file := `
name: test
version: "1.0.0"
language: go
targets:
  build:
    steps:
    - runner: go
      config:
        a: 3
`
	f := strings.NewReader(file)
	c, e := config.LoadFromReader[Config](f)

	require.NoError(t, e)

	assert.NotNil(t, c.TargetByName("build").Steps[0].ConfigRaw.Unmarshal)

	// Decode some special config.
	type C struct {
		a int `yaml:"a"`
	}
	cc := C{a: 1}
	require.NoError(t, c.TargetByID("test::build").Steps[0].ConfigRaw.Unmarshal(&cc))
}
