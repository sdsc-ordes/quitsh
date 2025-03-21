package common

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvTypeUnmarshal(t *testing.T) {
	type A struct {
		Env EnvironmentType
	}

	for _, k := range []EnvironmentType{EnvironmentDev, EnvironmentTesting, EnvironmentStaging, EnvironmentProd} {
		a := A{Env: k}
		s, e := yaml.Marshal(&a)
		require.NoError(t, e)

		b := A{Env: -1}
		e = yaml.Unmarshal(s, &b)
		require.NoError(t, e)

		assert.Equal(t, a, b)
	}
}
