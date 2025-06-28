package image

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageTypeUnmarshal(t *testing.T) {
	t.Parallel()

	type A struct {
		ImgType Type
	}

	for _, k := range GetAllImageTypes() {
		a := A{ImgType: k}
		s, e := yaml.Marshal(&a)
		require.NoError(t, e)

		b := A{ImgType: -1}
		e = yaml.Unmarshal(s, &b)
		require.NoError(t, e)

		assert.Equal(t, a, b)
	}
}
