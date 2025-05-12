package common

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTypeUnmarshal(t *testing.T) {
	t.Parallel()
	type A struct {
		BuildType BuildType
	}

	for _, k := range GetAllBuildTypes() {
		a := A{BuildType: k}
		s, e := yaml.Marshal(&a)
		require.NoError(t, e)

		b := A{BuildType: -1}
		e = yaml.Unmarshal(s, &b)
		require.NoError(t, e)

		assert.Equal(t, a, b)
	}
}
