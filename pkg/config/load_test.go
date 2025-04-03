package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type configD struct {
	V string `yaml:"value" default:"defaulted"`
}

type config struct {
	C string `yaml:"color" default:"defaulted"`
	V int    `yaml:"value" default:"99"`
	F string `yaml:"undefaulted"`

	D configD `yaml:"data"`
}

func (c *config) Init() error {
	if c.C == "blue" {
		c.V = 101
	}

	return nil
}

func TestLoadOverViper(t *testing.T) {
	{
		data := `
color: "bright"
undefaulted: "from-config"
`
		var c config
		reader := strings.NewReader(data)
		e := loadConfigs(&c, reader)
		require.NoError(t, e)
		require.Equal(t, "bright", c.C, "Should be set.")
		require.Equal(t, 99, c.V, "Should be defaulted.")
		require.Equal(t, "from-config", c.F, "Should be set by config.")
		require.Equal(t, "defaulted", c.D.V, "Should be defaulted.")
	}

	{
		data := `color: "red"`
		var c config
		reader := strings.NewReader(data)
		os.Setenv("QUITSH_COLOR", "blue")
		os.Setenv("QUITSH_DATA_VALUE", "banana")
		os.Setenv("QUITSH_UNDEFAULTED", "from-env")
		e := loadConfigs(&c, reader)
		require.NoError(t, e)
		require.Equal(t, "blue", c.C, "Should be set.")
		require.Equal(t, 101, c.V, "Should be set by init.")
		require.Equal(t, "from-env", c.F, "Should be set by env.")
		require.Equal(t, "banana", c.D.V, "Should be set by env.")
	}
}
