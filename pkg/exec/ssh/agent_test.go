//go:build test && (test_small || test_all)

package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent(t *testing.T) {
	socketPath, err := StartAgent()
	require.NoError(t, err)
	assert.NotEmpty(t, socketPath)
}
