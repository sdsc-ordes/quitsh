package pipeline

import (
	"bytes"
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ciAttrs struct {
	N int `yaml:"n"`
}

type ciCustom struct {
	C int `yaml:"c"`
}

// Interface for CIAttributes.
func (c *ciAttrs) Unmarshal(unmarshal func(any) error) error {
	return unmarshal(c)
}

// Interface for CIAttributes.
func (c *ciAttrs) Merge(other PipelineAttributes) {
	o := common.Cast[*ciAttrs](other)
	c.N += o.N
}

// Interface for CIAttributes.
func (c *ciAttrs) Clone() PipelineAttributes {
	var res = *c

	return &res
}

type All struct {
	General PipelineSettings
	Attrs   ciAttrs
	Custom  ciCustom
}

func (*All) Init() error { return nil }

func TestPipelineSettings(t *testing.T) {
	// Marshal
	all := All{
		General: PipelineSettings{Type: TagPipeline},
		Attrs:   ciAttrs{N: 4},
		Custom:  ciCustom{C: 10},
	}

	w := bytes.NewBuffer(nil)
	err := WritePipelineSettings(&all, w)
	require.NoError(t, err)
	t.Logf("File:\n---\n%v\n---", w.String())

	r := bytes.NewReader(w.Bytes())

	// Unmarshal
	all2, err := NewPipelineSettingsFromReader[All](r)
	t.Logf("Settings: %v", all2)
	require.NoError(t, err)

	assert.Equal(t, all, all2)
}
