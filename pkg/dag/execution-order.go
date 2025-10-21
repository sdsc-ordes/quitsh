package dag

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/common/recache"
	"github.com/sdsc-ordes/quitsh/pkg/common/set"
	"github.com/sdsc-ordes/quitsh/pkg/common/stack"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/input"
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"

	"deedles.dev/xiter"
)

// DefineExecutionOrderStage does the same as `defineExecutionOrder` but by selecting
// from `selection` only the targets which match `stage`.
// The argument `selection` can be `nil`.
func DefineExecutionOrderStage(
	all []*component.Component,
	selection []*component.Component,
	stage stage.Stage,
	inputPathChanges []string,
	rootDir string,
) (TargetNodeMap, Priorities, error) {
	log.Info("Define execution order.")

	sel := &TargetSelection{}

	debug.Assert(len(selection) <= len(all), "selection is longer than all nodes (?)")

	for i := range selection {
		c := selection[i].Config()
		for _, target := range c.Targets {
			if target.Stage != stage {
				continue
			}
			debug.Assert(!sel.Exists(target.ID), "target id '%v' should not exists", target.ID)
			sel.Insert(target.ID)
		}
	}

	nodes, prios, err := DefineExecutionOrder(all, sel, inputPathChanges, rootDir)

	if err != nil {
		log.ErrorE(err,
			"The direct acyclic graph defined by your components has errors:\n"+
				"- Check that you loaded enough components for the graph to be complete:\n"+
				"  -> working directory (or '-C') might be at the wrong place (use the top-level to check)\n"+
				"- Check that target ids and input ids are all correct."+
				"- Check that the graph has no cycles.",
		)

		return nil, nil, err
	}

	log.Info(prios.Format())

	return nodes, prios, err
}

// DefineExecutionOrder defines the execution order over all components.
// The target selection defines which target nodes are of interest, if `nil`
// all target leaf nodes are taken as selection.
// The `inputPathChanges` denote the path changes which will propagate the
// changed flags on the target nodes. If its `nil` all targets
// are considered changed by default.
// The `rootDir` is the `BaseDir` if `input.Config.RelativeToRoot` is used.
func DefineExecutionOrder(
	components []*component.Component,
	targetSelection *TargetSelection,
	inputPathChanges []string,
	rootDir string,
) (targets TargetNodeMap, prios Priorities, err error) {
	if targetSelection != nil && targetSelection.Len() == 0 {
		return nil, nil, nil
	}

	if targetSelection != nil {
		log.Debug("Defined selection.", "selection", targetSelection)
	}

	regexCache := recache.NewCache(true)

	log.Debug("Setup graph.")
	allNodes, allInputs, allComps, err := constructNodes(components, targetSelection, rootDir)
	if err != nil {
		return
	}

	graph, err := newGraph(allNodes, targetSelection)
	if err != nil {
		return
	}

	err = graph.SolveExecutionOrder()
	if err != nil {
		return
	}

	// Make all input path changes absolute.
	for i := range inputPathChanges {
		inputPathChanges[i] = fs.MakeAbsoluteTo(rootDir, inputPathChanges[i])
	}
	log.Debug("Changed paths.", "paths", inputPathChanges)

	err = graph.SolveInputChanges(allInputs, allComps, &regexCache, inputPathChanges)
	if err != nil {
		return
	}

	targets, prios = graph.NodesToPriorityList()

	return
}

func constructNodes(
	components []*component.Component,
	targetSelection *TargetSelection,
	rootDir string,
) (TargetNodeMap, map[input.ID]*input.Config, map[string]*component.Component, error) {
	allNodes := make(TargetNodeMap, len(components)*4) //nolint:mnd // intentional.
	allInputs := make(map[input.ID]*input.Config, len(components))
	allComps := make(map[string]*component.Component, len(components))

	// Add all components we found as nodes to the graph.
	for _, c := range components {
		config := c.Config()
		allComps[c.Name()] = c

		// Add all regexes to the cache.
		for _, in := range config.Inputs {
			// Set absolute base directory for all inputs changesets.
			if in.RelativeToRoot {
				in.BaseDir = rootDir
			} else {
				in.BaseDir = c.Root()
			}
			allInputs[in.ID] = in
		}

		for _, t := range config.Targets {
			if _, exists := allNodes[t.ID]; exists {
				debug.Assert(
					!exists,
					"target id '%s' in component '%s' already exists (must be unique)",
					t.ID,
					config.Name,
				)
			}

			tNode := &TargetNode{Target: t, Comp: c, Config: config}

			// Resolve target ids.
			resolveTargetIDs(tNode)

			allNodes[t.ID] = tNode
		}
	}

	log.Debug("Connect all target nodes.")
	allNodes, err := connectNodes(allNodes, targetSelection)
	if err != nil {
		return nil, nil, nil, err
	}

	// Resolve inputs over all nodes on the graph.
	for _, n := range allNodes {
		e := resolveInputIDs(n, allInputs, allComps)
		if e != nil {
			return nil, nil, nil, e
		}
	}

	return allNodes, allInputs, allComps, nil
}

type graph struct {
	nodes         TargetNodeMap
	execRootNodes []*TargetNode // at the start of the execution order.
	execLeafNodes []*TargetNode // at the end of the execution order.

	/// The selection subgraph of targets we want to solve for.
	/// =======================================================
	// All nodes in the selection, can be empty == no selection.
	nodesSel *set.Unordered[target.ID]
	// The leaf nodes of the subgraph (defaults points to `execLeafNodes`).
	execLeafNodesSel *[]*TargetNode
	// The root nodes of the subgraph (where we start executing,
	// defaults points to `execRootNodes`).
	execRootNodesSel *[]*TargetNode
	/// =======================================================
}

func (g *graph) recomputeSubgraph(selection *TargetSelection) error {
	if selection == nil {
		return nil
	}

	g.execLeafNodesSel = &[]*TargetNode{}
	g.execRootNodesSel = &[]*TargetNode{}

	for _, n := range g.nodes {
		if selection.Exists(n.Target.ID) {
			selection.Remove(n.Target.ID)
			*g.execLeafNodesSel = append(*g.execLeafNodesSel, n)
		}
	}

	if selection.Len() != 0 {
		return errors.New(
			"not all target ids from selection were found loaded components: remaining target ids: '%v'",
			selection,
		)
	}

	nodesSelection := set.NewUnorderedWithCap[target.ID](len(g.nodes))

	execRootNodesSelection := filterNodesBFS(
		*g.execLeafNodesSel,
		func(n *TargetNode) bool {
			nodesSelection.Insert(n.Target.ID)

			return len(n.Backward) == 0
		},
		backwardDir,
	)
	g.execRootNodesSel = &execRootNodesSelection
	g.nodesSel = &nodesSelection

	debug.Assert(
		len(*g.execLeafNodesSel) == 0 || len(*g.execRootNodesSel) != 0,
		"we found no root nodes from non-empty selection (this is a bug)",
	)

	return nil
}

func newGraph(
	nodes TargetNodeMap,
	selection *TargetSelection) (graph, error) {
	if selection != nil && selection.Len() == 0 {
		return graph{}, errors.New("graph selection must contain elements if not nil")
	}

	// Find all execution leaf nodes. (the ones with no children)
	execLeafNodes := []*TargetNode{}
	// Find all execution roots nodes.
	// (the one first in the execution order, the ones with no dependencies)
	execRootNodes := []*TargetNode{}

	// By default take all leaf nodes as selection.
	execLeafNodesSel := &execLeafNodes
	// By default take all root nodes as selection.
	execRootNodesSel := &execRootNodes

	for _, n := range nodes {
		if len(n.Forward) == 0 {
			execLeafNodes = append(execLeafNodes, n)
		}

		if len(n.Backward) == 0 {
			execRootNodes = append(execRootNodes, n)
		}
	}

	g := graph{
		nodes:            nodes,
		execRootNodes:    execRootNodes,
		execLeafNodes:    execLeafNodes,
		execLeafNodesSel: execLeafNodesSel,
		execRootNodesSel: execRootNodesSel,
	}

	err := g.CheckNoCycles()
	if err != nil {
		return g, err
	}

	err = g.recomputeSubgraph(selection)
	if err != nil {
		return g, err
	}

	printInfo(&g)

	return g, nil
}

func printInfo(g *graph) {
	log.Debug(
		"Graph infos.",
		"nodes",
		len(g.nodes),
		"roots",
		len(g.execRootNodes),
		"leaves",
		len(g.execLeafNodes),
	)

	if g.nodesSel != nil {
		log.Debug("Graph selection info.",
			"roots-selection",
			len(*g.execRootNodesSel),
			"leaves-selection",
			len(*g.execLeafNodesSel),
		)
	}
}

type traverseDirection int

const backwardDir traverseDirection = 0
const forwardDir traverseDirection = 1

// filterNodesBFS filters the graph in breadth-first manner with a `filter` function.
func filterNodesBFS(
	startNodes []*TargetNode,
	filter func(n *TargetNode) bool,
	dir traverseDirection,
) (result []*TargetNode) {
	visitNodesBFS(startNodes, func(n *TargetNode) bool {
		if filter(n) {
			result = append(result, n)
		}

		return true
	}, dir)

	return
}

// visitNodesBFS iterates over the graph in breadth-first manner.
func visitNodesBFS(
	startNodes []*TargetNode,
	visit func(n *TargetNode) bool,
	dir traverseDirection,
) {
	bfsStack := stack.NewStackWithCap[*TargetNode](len(startNodes))
	bfsStack.Push(startNodes...)

	for bfsStack.Len() != 0 {
		log.Tracef("Current BFS stack:\n%v", formatStack(&bfsStack, true))

		n := bfsStack.PopFront()
		if !visit(n) {
			return
		}

		switch dir {
		case backwardDir:
			bfsStack.Push(n.Backward...)
		case forwardDir:
			bfsStack.Push(n.Forward...)
		default:
			panic("wrong direction")
		}
	}
}

// resolveInputIDs resolves all `self|.*::XXX` in input ids in `.Inputs`.
func resolveInputIDs(
	node *TargetNode,
	allInputs map[input.ID]*input.Config,
	allComps map[string]*component.Component,
) error {
	for idx, inputID := range node.Target.Inputs {
		// Mangle `self` (referring to whole comp.)
		if inputID == "self" {
			inputID = input.DefineIDComp(node.Comp.Name())
			node.Target.Inputs[idx] = inputID
		}

		// Mangle `self::` into own components input id.
		trimmedID := strings.TrimPrefix(string(inputID), "self::")
		if trimmedID != string(inputID) {
			inputID = input.DefineID(node.Config.Name, trimmedID)
			node.Target.Inputs[idx] = inputID
		}

		if inputID.IsComponent() {
			if _, exists := allComps[string(inputID)]; !exists {
				return errors.New(
					"input id '%v' referring to a component in target '%v' on component '%v' is "+
						"not found on all found components (maybe search dir. wrong?)",
					inputID,
					node.Target.ID,
					node.Config.Name,
				)
			}
		} else {
			if _, exists := allInputs[inputID]; !exists {
				return errors.New(
					"input id '%v' in target '%v' on component '%v' is "+
						"not found on all found components (maybe search dir. wrong?)",
					inputID,
					node.Target.ID,
					node.Config.Name,
				)
			}
		}
	}

	return nil
}

// resolveTargetIDs resolves all `self::XXX` target ids in `.Dependencies`.
func resolveTargetIDs(node *TargetNode) {
	for idx, targetID := range node.Target.Dependencies {
		// Mangle `self::` into own components input id.
		trimmedID := strings.TrimPrefix(string(targetID), "self::")

		if trimmedID != string(targetID) {
			targetID = target.DefineID(node.Config.Name, trimmedID)
			node.Target.Dependencies[idx] = targetID
		}
	}
}

func connectNodes(nodes TargetNodeMap, sel *TargetSelection) (TargetNodeMap, error) {
	visited := make(map[target.ID]*TargetNode, len(nodes))

	dfsStack := stack.NewStackWithCap[target.ID](len(nodes))
	if sel != nil {
		for id := range sel.Values() {
			dfsStack.Push(id)
		}
	} else {
		for id := range nodes {
			dfsStack.Push(id)
		}
	}

	// Depth-first traversal, visit each node only once.
	for dfsStack.Len() != 0 {
		id := dfsStack.Pop()
		if _, ok := visited[id]; ok {
			continue
		}

		n, ok := nodes[id]
		if !ok {
			return nil, errors.New(
				"reached target '%s' started from target selection could "+
					"not be found by the loaded components", id)
		}

		visited[id] = n // Mark node as reached.

		// Connect the node (dependencies).
		visitedDeps := set.NewUnordered[target.ID]()
		for _, depName := range n.Target.Dependencies {
			if visitedDeps.Exists(depName) {
				// ignore multiple same dependencies.
				continue
			}
			visitedDeps.Insert(depName)

			depNode, exists := nodes[depName]
			if !exists {
				return nil, errors.New(
					"dependency target id '%s' defined on target '%s' does not exist",
					depName,
					n.Target.ID,
				)
			}

			// Add the backward node.
			n.Backward = append(n.Backward, depNode)

			// Add the forward node on the parent.
			depNode.Forward = append(depNode.Forward, n)
		}

		// Visit the dependencies.
		dfsStack.Push(n.Target.Dependencies...)
	}

	return visited, nil
}

// CheckNoCycles checks that the graph has no cycles by traversing it in
// depth-first manner from all roots.
func (graph *graph) CheckNoCycles() (err error) {
	log.Debug("Check for cycles.")

	startNodes := slices.Collect(maps.Values(graph.nodes))
	visited := set.NewUnordered[target.ID]()

	// Check for every start node if we run into a cycle.
	for _, root := range startNodes {
		if visited.Exists(root.Target.ID) {
			// DFS cycle search was already on this path, so skip it.
			continue
		}

		idsOnPath := make(map[target.ID]struct{})
		dfsStack := stack.NewStack[*TargetNode]()
		pathStack := stack.NewStack[target.ID]()

		stepDown := func(id target.ID) {
			pathStack.Push(id)
			idsOnPath[id] = struct{}{}
		}

		stepUpwards := func(id target.ID) {
			pathStack.Pop()
			delete(idsOnPath, id)
		}

		log.Trace("Start DFS cycle check: %v", root.Target.ID)
		dfsStack.Push(root)

		for dfsStack.Len() != 0 {
			currNode := dfsStack.Pop()
			currID := currNode.Target.ID

			visited.Insert(currID)

			if _, exists := idsOnPath[currID]; exists {
				return errors.Combine(
					err,
					errors.New(
						"direct acyclic graph contains a cycle in the following target chain:\n%v",
						formatPath(&pathStack),
					),
				)
			}

			stepDown(currID)
			log.Trace("Current DFS path:\n%v", formatPath(&pathStack))

			if len(currNode.Backward) == 0 {
				stepUpwards(currID)
			} else {
				dfsStack.Push(currNode.Backward...)
			}
		}
	}

	return err
}

func (graph *graph) SolveExecutionOrder() error {
	log.Debug("Solve execution order.")

	visitNodesBFS(
		*graph.execLeafNodesSel,
		func(n *TargetNode) bool {
			// The nodes we depend on must have a priority strictly higher than the current node.
			for _, d := range n.Backward {
				d.Priority = max(n.Priority+1, d.Priority)
			}

			return true
		},
		backwardDir)

	return nil
}

// Propagates all input changes from all execution root nodes.
// to determine what changed. If the `paths` is nil
// all nodes are treated as changed.
// This function does a forward traversal up the tree to determine the changed status.
// Afterwards, the subgraph selection in `graph` is recomputed, defined by all
// visited nodes from a backwards traversal starting from all **changed** nodes.
func (graph *graph) SolveInputChanges(
	inputs map[input.ID]*input.Config,
	comps map[string]*component.Component,
	regexCache *recache.Cache,
	paths []string) error {
	log.Debug("Solve input changes.")

	var changed TargetSelection
	inputChanges := make(map[input.ID]InputChanges, len(inputs))

	// Run the graph upwards (in execution order) and
	// propagate input changes and determine if
	// target is changed or not.
	for _, root := range *graph.execRootNodesSel {
		dfsStack := stack.NewStack[*TargetNode]()
		dfsStack.Push(root)

		for dfsStack.Len() != 0 {
			log.Trace("Current DFS stack:\n%v", formatStack(&dfsStack, false))

			n := dfsStack.Pop()

			if !graph.inSelection(n) {
				log.Trace("Skipped node not in selection.")

				continue
			}

			currIn := &n.Inputs
			if currIn.ChangedByDependency {
				debug.Assert(
					currIn.Paths == nil,
					"we should not have changed paths, when we skipped detection on this target",
					"paths",
					currIn.Paths,
				)
				debug.Assert(currIn.Changed, "we should have changes when skipped it set")
			}

			switch {
			case paths == nil:
				log.Trace(
					"No paths given -> setting target id '%v' to changed.",
					n.Target.ID,
				)
				currIn.Changed = true

			case !currIn.ChangedByDependency:
				log.Trace("Detect change for current node '%v'.", n.Target.ID)
				changed, changes, err := determineChangedPaths(
					n,
					inputs,
					comps,
					inputChanges,
					regexCache,
					paths,
				)
				if err != nil {
					return err
				}

				currIn.Changed = changed
				currIn.Paths = changes

			default:
				log.Trace("Current target '%s' already changed.", n.Target.ID)
				// we are skipped
			}

			log.Debug(
				"Changes for target id.",
				"id",
				n.Target.ID.String(),
				"changed",
				currIn.Changed,
			)

			if currIn.Changed {
				changed.Insert(n.Target.ID)
			}

			// Propagate to children, by merging the input changes.
			for _, c := range n.Forward {
				c.Inputs.Merge(&n.Inputs)
			}

			// Go to next nodes.
			dfsStack.Push(n.Forward...)
		}
	}

	return graph.recomputeSubgraph(&changed)
}

// NodesToPriorityList converts the priorities on the nodes to an
// ordered list in descending order and returns all targets.
// Only changed targets are returned.
func (graph *graph) NodesToPriorityList() (nodes TargetNodeMap, result Priorities) {
	nodes = make(TargetNodeMap)
	prios := make(map[int]PrioritySet)

	log.Debug("Construct priority set.")
	visited := set.NewUnorderedWithCap[target.ID](len(graph.nodes))

	visitNodesBFS(
		*graph.execRootNodesSel,
		func(n *TargetNode) bool {
			if !n.Inputs.Changed {
				return true
			}

			if !graph.inSelection(n) || visited.Exists(n.Target.ID) {
				return true
			}

			log.Tracef("Adding node '%v' with prio '%v'.", n.Target.ID, n.Priority)

			// Add the node to the result.
			prio := prios[n.Priority]
			prios[n.Priority] = PrioritySet{
				Priority: n.Priority,
				Nodes:    append(prio.Nodes, n),
			}
			nodes[n.Target.ID] = n

			visited.Insert(n.Target.ID)

			return true
		},
		forwardDir)

	// Sort all priorities in descending order.
	for x := range slices.Backward(slices.Collect(xiter.Sorted(maps.Keys(prios)))) {
		result = append(result, prios[x])
	}

	return
}

// inSelection determines in case of a selection if the current node is
// inside the selected subgraph.
func (graph *graph) inSelection(n *TargetNode) bool {
	return graph.nodesSel == nil ||
		graph.nodesSel.Exists(n.Target.ID)
}

func determineChangedPathsDefault(
	rootDir string,
	paths []string,
) (changed bool, changes []string) {
	log.Trace("Check for changes in dir '%v'", rootDir)
	for i := range paths {
		debug.Assert(path.IsAbs(paths[i]), "input path '%s' must be absolute", paths[i])
		if _, changed = input.BaseDir(rootDir).TrimOffFrom(paths[i]); changed {
			changes = append(changes, paths[i])

			return
		}
	}

	return
}

// determineChangedPaths determines if this target has changes on its own.
//
//nolint:gocognit
func determineChangedPaths(
	node *TargetNode,
	inputs map[input.ID]*input.Config,
	comps map[string]*component.Component,
	inputChanges map[input.ID]InputChanges,
	regexCache *recache.Cache,
	paths []string,
) (bool, []string, error) {
	if node.Target.Inputs == nil {
		// If there are no inputs, check if just any relative path match,
		// thats the default rule.
		changed, changes := determineChangedPathsDefault(node.Comp.Root(), paths)

		return changed, changes, nil
	}

	// Else check all input rules.
	for _, inputID := range node.Target.Inputs {
		// If the input id refers to a components, check against the root directory.
		if inputID.IsComponent() {
			comp := comps[string(inputID)]
			debug.Assert(comp != nil, "component referred is not existing")
			changed, changes := determineChangedPathsDefault(comp.Root(), paths)

			return changed, changes, nil
		}

		// otherwise check the input set
		log.Trace("Check for changes for input id '%v'", inputID)
		input, exists := inputs[inputID]
		debug.Assert(exists, "input id '%s' does not exist", inputID)

		inputCh := inputChanges[inputID]

		if inputCh.Changed {
			// Input changeset already determined.
			return inputCh.Changed, inputCh.Changes, nil
		}

		var relativePaths []string
		for i := range paths {
			if rel, matches := input.TrimOfBaseDir(paths[i]); matches {
				relativePaths = append(relativePaths, rel)
			}
		}

		if len(relativePaths) == 0 {
			// no paths which match this input set.
			log.Trace("Relative paths do not match.")

			continue
		}

		includeRegexes, e := regexCache.Get(input.Includes()...)
		if e != nil {
			return false, nil, errors.AddContext(
				e,
				"failed to get include regexes in target id '%s' for input id '%s'",
				node.Target.ID,
				inputID,
			)
		}

		excludeRegexes, e := regexCache.Get(input.Excludes()...)
		if e != nil {
			return false, nil, errors.AddContext(
				e,
				"failed to get exclude regexes in target id '%s' for input id '%s'",
				node.Target.ID,
				inputID,
			)
		}

		for _, p := range relativePaths {
			inputCh.Changed = includeRegexes.Match(p) && !excludeRegexes.Match(p)

			if inputCh.Changed {
				log.Trace("Matched path '%v'.", p)

				inputCh.Changes = append(inputCh.Changes, p)
				inputChanges[inputID] = inputCh // set the changes back!

				return inputCh.Changed, inputCh.Changes, nil
			}
		}
	}

	return false, nil, nil
}

// formatPath formats the current stack with some
// nice output for the log.
func formatPath(stack *stack.Stack[target.ID]) string {
	var sb strings.Builder

	stack.VisitUpward(func(i int, n target.ID) bool {
		sb.WriteString(fmt.Sprintf("  - '%v'\n", n))
		if i != stack.Len()-1 {
			sb.WriteString("  |\n  v\n")
		}

		return true
	})

	return sb.String()
}

func formatStack(stack *stack.Stack[*TargetNode], withPrio bool) string {
	var sb strings.Builder

	stack.VisitUpward(func(_ int, n *TargetNode) bool {
		sb.WriteString(fmt.Sprintf("  - '%v'", n.Target.ID))
		if withPrio {
			sb.WriteString(fmt.Sprintf(" prio: '%v'", n.Priority))
		}
		sb.WriteString("\n")

		return true
	})

	return sb.String()
}

type TargetNodeMap = map[target.ID]*TargetNode
type TargetSelection = set.Unordered[target.ID]

type PrioritySet struct {
	Priority int
	Nodes    []*TargetNode
}

type Priorities []PrioritySet

// Format formats the priorities.
func (prios *Priorities) Format() string {
	var sb strings.Builder
	sb.WriteString("Execution Targets:\n")
	for _, set := range *prios {
		sb.WriteString(fmt.Sprintf("- Priority: '%v'\n", set.Priority))
		for _, n := range set.Nodes {
			sb.WriteString(fmt.Sprintf("  - '%v'\n", n.Target.ID))
		}
	}

	return sb.String()
}
