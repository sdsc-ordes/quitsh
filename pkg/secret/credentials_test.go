package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentials(t *testing.T) {
	env := []string{"AAA=pass", "BBB=gabriel"}

	c := CredentialsEnv{TokenEnv: "AAA", UserEnv: "BBB"}
	cc, err := c.ResolveFrom(false, env)

	require.NoError(t, err)
	assert.Equal(t, "pass", cc.Token())
	assert.Equal(t, "gabriel", cc.User())
}
