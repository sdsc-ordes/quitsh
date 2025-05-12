package stage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStagePriority(t *testing.T) {
	t.Parallel()
	stages := NewDefaults()

	assert.True(t, stages[0].IsBefore(stages[1]))
	assert.False(t, stages[0].IsBefore(stages[2]))

	assert.True(t, stages[1].IsAfter(stages[0]))
	assert.False(t, stages[2].IsAfter(stages[0]))
}
