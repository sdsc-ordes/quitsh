package common

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Custom struct {
	Check bool
}

var unmarshalCalled = false //nolint:gochecknoglobals // intentional.
var marshalCalled = false   //nolint:gochecknoglobals // intentional.

func (b *Custom) UnmarshalYAML(unmarshal func(any) error) error {
	unmarshalCalled = true

	return unmarshal(&b.Check)
}

func (b Custom) MarshalYAML() (any, error) {
	marshalCalled = true

	return &b.Check, nil
}

type A struct {
	Value Custom `yaml:"value"`
}

func TestGoccyYAML(t *testing.T) {
	a := A{Value: Custom{true}}
	s, e := yaml.Marshal(&a)
	t.Logf("Yaml:\n---\n%s\n---", string(s))

	assert.True(t, marshalCalled, "mashall not called")
	require.NoError(t, e)

	b := A{}
	e = yaml.Unmarshal(s, &b)
	assert.True(t, unmarshalCalled, "unmarshal not called")
	require.NoError(t, e)

	assert.Equal(t, a, b)
}
