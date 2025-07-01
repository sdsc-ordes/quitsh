package build

import (
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestVersionCorrect(t *testing.T) {
	r, e := os.ReadFile("../../.component.yaml")
	require.NoError(t, e)

	type V struct {
		Version string `yaml:"version"`
	}
	var v V
	e = yaml.Unmarshal(r, &v)
	require.NoError(t, e)

	require.Equal(t, v.Version, GetBuildVersion().String(),
		"you must update the version.go file correctly, use .githooks/pre-commit.")
}
