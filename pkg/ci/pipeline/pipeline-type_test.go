package pipeline

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineTypeUnmarshal(t *testing.T) {
	t.Parallel()
	type A struct {
		PT PipelineType `yaml:"pt"`
	}

	for _, k := range GetAllPipelineTypes() {
		a := A{PT: k}
		s, e := yaml.Marshal(&a)
		require.NoError(t, e)

		b := A{PT: -1}
		e = yaml.Unmarshal(s, &b)
		require.NoError(t, e)

		assert.Equal(t, a, b)
	}
}

func TestPipelineTypeUnmarshal2(t *testing.T) {
	t.Parallel()
	type A struct {
		PT PipelineType `yaml:"pt"`
	}

	s := `pt: merge-request`

	pt, e := NewPipelineType("merge-request")
	require.NoError(t, e)
	a := A{PT: pt}

	b := A{PT: -1}
	e = yaml.Unmarshal([]byte(s), &b)
	require.NoError(t, e)
	assert.Equal(t, a, b)
}
