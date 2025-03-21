package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvListFind(t *testing.T) {
	env := EnvList{"a=fff", "b=ggg"}
	val := env.FindIdx("a")
	assert.Equal(t, "fff", val.Value)
	assert.Equal(t, 0, val.Idx())
	assert.True(t, val.Defined())

	val = env.FindIdx("b")
	assert.Equal(t, "ggg", val.Value)
	assert.Equal(t, 1, val.Idx())
	assert.True(t, val.Defined())

	val = env.FindIdx("c")
	assert.Equal(t, "", val.Value)
	assert.Equal(t, -1, val.Idx())
	assert.False(t, val.Defined())

	assert.Equal(t, "fff", env.Find("a"))

	env = EnvList(os.Environ())
	assert.NotEmpty(t, env.Find("PATH"))
}

func TestEnvListFindAll(t *testing.T) {
	env := EnvList{"a=  fff", "b=ggg", "c=ttt  ", "dummy"}
	m := env.FindAll("a", "c", "dummy")
	assert.Len(t, m, 3)
	assert.Equal(t, EnvValue{"  fff", 1}, m["a"])
	assert.Equal(t, EnvValue{"ttt  ", 3}, m["c"])
	assert.Equal(t, EnvValue{"", 4}, m["dummy"])

	m = env.FindAll("a", "c", "dummy2")
	assert.Len(t, m, 3)
	assert.Equal(t, EnvValue{"  fff", 1}, m["a"])
	assert.Equal(t, EnvValue{"ttt  ", 3}, m["c"])
	assert.Equal(t, EnvValue{}, m["dummy2"])
}

func TestEnvMap(t *testing.T) {
	env := EnvList{"a=  fff", "b=ggg", "c=ttt  ", "dummy"}
	m := NewEnvMapFromList(env)
	assert.Len(t, m, 4)
	assert.Equal(t, EnvValue{"  fff", 1}, m["a"])
	assert.Equal(t, EnvValue{"ggg", 2}, m["b"])
	assert.Equal(t, EnvValue{"ttt  ", 3}, m["c"])
	assert.Equal(t, EnvValue{"", 4}, m["dummy"])
	assert.Equal(t, EnvValue{}, m["dummy2"])
}

func TestEnvMapDup(t *testing.T) {
	env := EnvList{"a=  fff", "b=ggg", "c=ttt  ", "a=dummy"}
	m := NewEnvMapFromList(env)
	assert.Len(t, m, 3)
	assert.Equal(t, EnvValue{"dummy", 4}, m["a"])
	assert.Equal(t, EnvValue{"ggg", 2}, m["b"])
	assert.Equal(t, EnvValue{"ttt  ", 3}, m["c"])
	assert.Equal(t, EnvValue{}, m["dummy2"])
}

//nolint:testifylint // intentional.
func TestEnvMapDefined(t *testing.T) {
	env := EnvList{"a=", "b=ggg", "c=ttt  "}
	res := env.FindAll("b", "c")
	assert.NoError(t, res.AssertDefined())
	assert.NoError(t, res.AssertNotEmpty())

	res = env.FindAll("a", "c")
	assert.NoError(t, res.AssertDefined())
	assert.Error(t, res.AssertNotEmpty(), res)

	res = env.FindAll("d", "dd")
	assert.Error(t, res.AssertDefined())
	assert.Error(t, res.AssertNotEmpty())
}
