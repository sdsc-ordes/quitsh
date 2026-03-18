//go:build test && (test_small || test_all)

package ssh

import (
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent(t *testing.T) {
	log := log.Global()
	agent, err := StartAgent(log, WithNewInstance(true))

	require.NoError(t, err)
	assert.NotEmpty(t, agent.SocketPath)
	assert.NotEmpty(t, agent.pid)

	err = agent.Close()
	require.NoError(t, err)
}
