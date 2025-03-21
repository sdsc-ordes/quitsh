package pipeline

import (
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Settings struct {
	BuildAll      bool     `yaml:"buildAll"`
	PushManifests bool     `yaml:"pushManifests"`
	Count         int      `yaml:"count"`
	Values        []string `yaml:"values"`
}

func (s Settings) Clone() PipelineAttributes {
	return &s
}

func (s *Settings) Merge(other PipelineAttributes) {
	o := common.Cast[*Settings](other)

	s.BuildAll = s.BuildAll || o.BuildAll
	s.PushManifests = s.PushManifests || o.PushManifests
	s.Count += o.Count
	s.Values = append(s.Values, o.Values...)
}

func TestParseCISettings(t *testing.T) {
	description := "```yaml {ci}\n" + `
buildAll: true
pushManifests: true
count: 3
values: ["banana", "monkey"]` + "\n```"

	t.Logf("desc: %s", description)

	sett := Settings{}
	err := ParseCIAttributes(&sett, description)
	require.NoError(t, err)

	assert.True(t, sett.BuildAll)
	assert.True(t, sett.PushManifests)
	assert.Equal(t, []string{"banana", "monkey"}, sett.Values)
}

func TestParseCISettingsMerge(t *testing.T) {
	description1 := "```yaml {ci}\n" + `
buildAll: true
pushManifests: false
count: 3
values: ["banana"]` + "\n```"

	description2 := "```yaml {ci}\n" + `
buildAll: false
pushManifests: true
count: 1
values: ["monkey"]` + "\n```"

	t.Logf("desc: %s", description1)
	t.Logf("desc: %s", description2)

	sett := Settings{}
	var s PipelineAttributes = &sett

	err := ParseCIAttributes(s, description1, description2)
	require.NoError(t, err)

	assert.True(t, sett.BuildAll)
	assert.True(t, sett.PushManifests)
	assert.Equal(t, 4, sett.Count)
	assert.Equal(t, []string{"banana", "monkey"}, sett.Values)
}

func TestParseCISettingsError(t *testing.T) {
	description := "```yaml {ci}\n" + `
buildAll: true
pushManifests: true
values: ["banana", "monkey"]`

	t.Logf("desc: %s", description)

	sett := Settings{}
	err := ParseCIAttributes(&sett, description)
	require.NoError(t, err)

	assert.False(t, sett.BuildAll)
	assert.False(t, sett.PushManifests)
	assert.Nil(t, sett.Values)
}

func TestParseCISettingsError2(t *testing.T) {
	description := "```yaml {ci}\n" + `
buildAll: true
pushMani` + "\n```\n" + `fests: true
values: ["banana", "monkey"]`

	t.Logf("desc: %s", description)

	sett := Settings{}
	err := ParseCIAttributes(&sett, description)
	require.Error(t, err)

	assert.False(t, sett.BuildAll)
	assert.False(t, sett.PushManifests)
	assert.Nil(t, sett.Values)
}
