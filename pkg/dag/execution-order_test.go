package dag

import (
	"testing"

	"github.com/sdsc-ordes/quitsh/pkg/common/set"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/input"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const rootDir = "/repo"

func TestGraphExecOrder3Comps(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	comps, _ := generate3Comps(t)
	_, prios, e := DefineExecutionOrder(comps, rootDir)
	require.NoError(t, e)

	testGenerate3Comps(t, prios, nil, false)
}

func TestGraphExecOrder3CompsDoNotResolveIds(t *testing.T) {
	// When no input changes given, ids should not be resolved cause not needed.
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	comps, _ := generate3Comps(t)
	tgt := comps[0].Config().Targets["build1"]
	tgt.Inputs = append(tgt.Inputs, "non-existing-id")

	_, prios, e := DefineExecutionOrder(comps, rootDir)
	require.NoError(t, e)

	testGenerate3Comps(t, prios, nil, false)
}

func TestGraphExecOrder3CompsWithChanges(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	comps, paths := generate3Comps(t)
	_, prios, e := DefineExecutionOrder(comps, rootDir, WithInputChanges(paths))
	require.NoError(t, e)

	testGenerate3Comps(t, prios, paths, false)
}

func TestGraphExecOrder3CompsNoFlakey(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	for range 100 {
		comps, paths := generate3Comps(t)
		_, prios, e := DefineExecutionOrder(comps, rootDir, WithInputChanges(paths))
		require.NoError(t, e)

		testGenerate3Comps(t, prios, paths, false)
	}
}

func TestGraphExecOrder3CompsSel(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	log.Info("Run test with a root selection -> must not change anything.")
	comps, paths := generate3Comps(t)
	sel := set.NewUnordered[target.ID]("3::build3")
	_, prios, e := DefineExecutionOrder(
		comps,
		rootDir,
		WithTargetSelection(&sel),
		WithInputChanges(paths),
	)
	require.NoError(t, e)

	testGenerate3Comps(t, prios, paths, false)
}

func TestGraphExecOrder3CompsSel2(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	log.Info("Run test with a root selection -> must not change anything.")
	comps, paths := generate3Comps(t)

	conf := &component.Config{
		Name:     "4",
		Language: "go",
		Targets: map[string]*target.Config{
			// Other stage which should not be found.
			"test": {
				Dependencies: []target.ID{"2::build2", "1::build1"},
			},
		},
	}
	err = conf.Init()
	require.NoError(t, err)
	comp4 := component.NewComponent(conf, "/repo/components/4", "", "")
	comps = append(comps, &comp4)

	_, prios, e := DefineExecutionOrder(
		comps,
		rootDir,
		WithTargetsByStageFromComponents(comps[len(comps)-2:], "build"),
		WithInputChanges(paths),
	)
	require.NoError(t, e)

	testGenerate3Comps(t, prios, paths, false)
}

func testGenerate3Comps(
	t *testing.T,
	prios Priorities,
	paths []string,
	onlyOneComp bool,
) {
	require.Len(t, prios, 3)

	pI := 0
	compName := "1"
	assert.Equal(t, 2, prios[pI].Priority)
	require.Len(t, prios[pI].Nodes, 1)
	assert.EqualValues(t, compName+"::build1", prios[pI].Nodes[0].Target.ID)
	n := prios[pI].Nodes[0]
	if paths != nil {
		assert.False(t, n.Inputs.Changed)
		assert.False(t, n.Inputs.ChangedByDependency)
	} else {
		assert.True(t, n.Inputs.Changed)
		assert.False(t, n.Inputs.ChangedByDependency)
	}

	if !onlyOneComp {
		compName = "2"
	}
	pI = 1
	n = prios[pI].Nodes[0]
	assert.Equal(t, 1, prios[pI].Priority)
	require.Len(t, prios[pI].Nodes, 1)
	assert.EqualValues(t, compName+"::build2", n.Target.ID)
	if paths != nil {
		assert.True(t, n.Inputs.Changed)
		assert.False(t, n.Inputs.ChangedByDependency)
		p, _ := input.BaseDir(rootDir).TrimOffFrom(paths[1])
		assert.Equal(t, p, prios[pI].Nodes[0].Inputs.Paths[0])
	} else {
		assert.True(t, n.Inputs.Changed)
		assert.True(t, n.Inputs.ChangedByDependency)
	}

	if !onlyOneComp {
		compName = "3"
	}
	pI = 2
	n = prios[pI].Nodes[0]
	assert.Equal(t, 0, prios[pI].Priority)
	require.Len(t, prios[pI].Nodes, 1)
	assert.EqualValues(t, compName+"::build3", n.Target.ID)
	if paths != nil {
		assert.False(t, n.Inputs.Changed)
		assert.True(t, n.Inputs.ChangedByDependency)
	} else {
		assert.True(t, n.Inputs.Changed)
		assert.True(t, n.Inputs.ChangedByDependency)
	}
}

func TestGraphExecOrder3CompsSimpleSel(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	comps, paths := generate3Comps(t)

	sel := set.NewUnordered[target.ID]("2::build2")
	_, prios, err := DefineExecutionOrder(
		comps,
		rootDir,
		WithTargetSelection(&sel),
		WithInputChanges(paths),
	)

	require.NoError(t, err)
	require.Len(t, prios, 2)

	pI := 0
	assert.Equal(t, 1, prios[pI].Priority)
	require.Len(t, prios[pI].Nodes, 1)
	assert.EqualValues(t, "1::build1", prios[pI].Nodes[0].Target.ID)
	assert.False(t, prios[pI].Nodes[0].Inputs.Changed)
	assert.False(t, prios[pI].Nodes[0].Inputs.ChangedByDependency)

	pI = 1
	assert.Equal(t, 0, prios[pI].Priority)
	require.Len(t, prios[pI].Nodes, 1)
	assert.EqualValues(t, "2::build2", prios[pI].Nodes[0].Target.ID)
	assert.True(t, prios[pI].Nodes[0].Inputs.Changed)
	assert.False(t, prios[pI].Nodes[0].Inputs.ChangedByDependency)
}

func TestGraphExecOrderOneComp(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	comps, paths := generateOneComp(t)
	_, prios, e := DefineExecutionOrder(comps, rootDir, WithInputChanges(paths))
	require.NoError(t, e)
	testGenerate3Comps(t, prios, paths, true)

	log.Info("Run test with selection -> must be the same.")
	comps, paths = generateOneComp(t)
	sel := set.NewUnordered[target.ID]("1::build3")
	_, prios, e = DefineExecutionOrder(
		comps,
		rootDir,
		WithTargetSelection(&sel),
		WithInputChanges(paths),
	)
	require.NoError(t, e)
	testGenerate3Comps(t, prios, paths, true)
}

func TestGraphExecOrderSimpleCycle1(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	// Add a cycle
	comps, paths := generate3Comps(t)
	d := &comps[0].Config().Targets["build1"].Dependencies
	*d = append(*d, "3::build3")
	validate(t, comps...)

	_, _, e := DefineExecutionOrder(comps, "/", WithInputChanges(paths))
	require.ErrorContains(t, e, "contains a cycle")
}

func TestGraphExecOrderSimpleCycle2(t *testing.T) {
	t.Parallel()
	err := log.SetLevel("trace")
	require.NoError(t, err)

	comps, paths := generate3Comps(t)
	d := &comps[0].Config().Targets["build1"].Dependencies
	*d = append(*d, "2::build2")

	_, _, e := DefineExecutionOrder(comps, "/", WithInputChanges(paths))
	require.ErrorContains(t, e, "contains a cycle")
}

func TestGraphExecOrderSimpleCycle3(t *testing.T) {
	t.Parallel()
	// Create a simple graph with 2 components and a cycle.
	// 1::image -> 2::lint -> 2::test -> 1::image
	//          -> 2::build
	// ^----------------------|

	comps := []*component.Component{}

	config := &component.Config{
		Name:     "1",
		Language: "go",
		Targets: map[string]*target.Config{
			"image": {
				Dependencies: []target.ID{"2::lint", "2::build"},
			},
		},
	}
	err := config.Init()
	require.NoError(t, err)
	comp1 := component.NewComponent(config, "/repo/components/1", "", "")

	config = &component.Config{
		Name:     "2",
		Language: "go",
		Targets: map[string]*target.Config{
			"test": {
				Dependencies: []target.ID{"1::image"}, // that creates a cycle.
			},
			"build": {},
			"lint": {
				Dependencies: []target.ID{"self::test"},
			},
		},
	}
	err = config.Init()
	require.NoError(t, err)
	comp2 := component.NewComponent(config, "/repo/components/2", "", "")

	comps = append(comps, &comp1, &comp2)

	_, _, e := DefineExecutionOrder(comps, rootDir)
	require.ErrorContains(t, e, "contains a cycle")
}

func TestGraphExecOrderSimpleNoConnection(t *testing.T) {
	t.Parallel()
	comps, paths := generate2CompsWithNoConn(t)

	rootDir := "/repo"
	targets, prios, e := DefineExecutionOrder(comps, rootDir, WithInputChanges(paths))
	require.NoError(t, e)

	assert.Len(t, targets, 2)
	_, exists := targets[comps[0].Config().TargetByID("1::build").ID]
	assert.True(t, exists)
	_, exists = targets[comps[1].Config().TargetByID("2::build").ID]
	assert.True(t, exists)

	// No connections, one parallel set.
	assert.Len(t, prios, 1)
}

func generate3Comps(t *testing.T) ([]*component.Component, []string) {
	// Create a simple graph:
	// 1 <- 2 <- 3
	// ^---------|
	// where only node 2 should propagate the changed flag to target 3.

	paths := []string{
		"/repo/components/1/a/b/c/file", // is ignored.

		"/repo/components/2/!-file-must-match", // must match.
		"/repo/components/3/a/b/c/file",        // is ignored.
	}

	comps := []*component.Component{}
	conf := &component.Config{
		Name:     "1",
		Language: "go",
		Inputs: map[string]*input.Config{
			"in1": {
				Patterns: []string{"!^.*/b/.*/file$"}, // ignore!
			},
		},
		Targets: map[string]*target.Config{
			"build1": {
				Stage:  "build",
				Inputs: []input.ID{"self::in1"},
			},
		},
	}
	err := conf.Init()
	require.NoError(t, err)
	comp1 := component.NewComponent(conf, "/repo/components/1", "", "")

	conf = &component.Config{
		Name:     "2",
		Language: "go",
		Inputs: map[string]*input.Config{
			"in2": {
				Patterns:       []string{"^components/.*/!-file-must-.*$"},
				RelativeToRoot: true,
			},
		},
		Targets: map[string]*target.Config{
			"build2": {
				Stage:        "build",
				Inputs:       []input.ID{"self::in2"},
				Dependencies: []target.ID{"1::build1"},
			},
		},
	}

	err = conf.Init()
	require.NoError(t, err)
	comp2 := component.NewComponent(conf, "/repo/components/2", "", "")

	conf = &component.Config{
		Name:     "3",
		Language: "go",
		Inputs: map[string]*input.Config{
			"in3": {
				Patterns: []string{"!^/.*/b/c/.*$"}, // ignore!
			},
		},
		Targets: map[string]*target.Config{
			"build3": {
				Stage:        "build",
				Inputs:       []input.ID{"self::in3"},
				Dependencies: []target.ID{"2::build2", "1::build1"},
			},
		},
	}
	err = conf.Init()
	require.NoError(t, err)
	comp3 := component.NewComponent(conf, "/repo/components/3", "", "")

	comps = append(comps, &comp1, &comp2, &comp3)
	validate(t, comps...)

	return comps, paths
}

func generateOneComp(t *testing.T) ([]*component.Component, []string) {
	// Create a simple graph with one component.
	// target-1 <- target-2 <- target-3
	// ^------------------------------|
	// where only node 2 should propagate the changed flag to target 3.

	paths := []string{
		"/repo/components/2/a/b/c/file",        // is ignored.
		"/repo/components/2/!-file-must-match", // must match.
		"/repo/components/2/a/b/c/file",        // is ignored.
	}

	comps := []*component.Component{}
	conf := &component.Config{
		Name:     "1",
		Language: "go",
		Inputs: map[string]*input.Config{
			"in1": {
				Patterns: []string{"!^.*/b/.*/file$"}, // ignore!
			},
			"in2-no-match": {
				Patterns:       []string{"non-match"},
				RelativeToRoot: true,
			},
			"in2-match": {
				Patterns:       []string{"^components/.*/!-file-must-.*$"},
				RelativeToRoot: true,
			},
			"in3": {
				Patterns: []string{"!^/.*/b/c/.*$"}, // ignore!
			},
		},
		Targets: map[string]*target.Config{
			"build1": {
				Inputs: []input.ID{"self::in1"},
			},
			"build2": {
				// ------------------------* does also not match
				Inputs:       []input.ID{"self", "self::in2-no-match", "self::in2-match"},
				Dependencies: []target.ID{"self::build1"},
			},
			"build3": {
				Inputs:       []input.ID{"self::in3"},
				Dependencies: []target.ID{"1::build2", "self::build1"},
			},
		},
	}
	err := conf.Init()
	require.NoError(t, err)
	comp1 := component.NewComponent(conf, "/repo/components/1", "", "")

	comps = append(comps, &comp1)
	validate(t, comps...)

	return comps, paths
}

func generate2CompsWithNoConn(t *testing.T) ([]*component.Component, []string) {
	// Create a simple graph with two component.
	// 1::build,  2::build
	// Where 2::build uses the same input changeset from `1::src` and should get changed.

	paths := []string{
		"/repo/components/1/src/file",
	}

	comps := []*component.Component{}
	conf1 := &component.Config{
		Name:     "1",
		Language: "go",
		Inputs: map[string]*input.Config{
			"src": {
				Patterns: []string{"^src/.*$"}, // ignore!
			},
		},
		Targets: map[string]*target.Config{
			"build": {
				Inputs: []input.ID{"self::src"},
			},
		},
	}
	err := conf1.Init()
	require.NoError(t, err)
	comp1 := component.NewComponent(conf1, "/repo/components/1", "", "")

	conf2 := &component.Config{
		Name:     "2",
		Language: "go",
		Targets: map[string]*target.Config{
			"build": {
				Inputs: []input.ID{"1::src"},
			},
		},
	}
	err = conf2.Init()
	require.NoError(t, err)
	comp2 := component.NewComponent(conf2, "/repo/components/2", "", "")

	comps = append(comps, &comp1, &comp2)
	validate(t, comps...)

	return comps, paths
}

func validate(t *testing.T, comps ...*component.Component) {
	for _, c := range comps {
		require.NoError(t, c.Config().Init())
	}
}
