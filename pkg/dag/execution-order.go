package dag

import (
	"fmt"
	"iter"
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

type (
	TargetNodeMap   = map[target.ID]*TargetNode
	TargetSelection = set.Unordered[target.ID]

	PrioritySet struct {
		Priority int
		Nodes    []*TargetNode
	}

	Priorities []PrioritySet

	graph struct {
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

	ExecOption func(*opts) error

	opts struct {
		targetSelection *TargetSelection

		nodeCount int

		inputPathChanges []string
	}
)

// DefineExecutionOrder defines the execution order over all components.
// The `rootDir` is the `BaseDir` if `input.Config.RelativeToRoot` is used.
// Options:
//   - The target selection defines which target nodes are of interest, if `nil`
//     all target leaf nodes are taken as selection.
//   - The `inputPathChanges` denote the path changes which will propagate the
//     changed flags on the target nodes. If its `nil` all targets
//     are considered changed by default.
func DefineExecutionOrder(
	components []*component.Component,
	rootDir string,
	option ...ExecOption,
) (TargetNodeMap, Priorities, error) {
	log.Info("Define execution order.")

	nodes, prios, err := defineExecutionOrder(components, rootDir, option...)

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

func defineExecutionOrder(
	components []*component.Component,
	rootDir string,
	options ...ExecOption,
) (targets TargetNodeMap, prios Priorities, err error) {
	o := opts{nodeCount: len(components)}
	err = o.Apply(options...)
	if err != nil {
		return nil, nil, err
	}

	if o.targetSelection != nil && o.targetSelection.Len() == 0 {
		return nil, nil, nil
	}

	if o.targetSelection != nil {
		log.Debug("Defined selection.", "selection", o.targetSelection)
	}

	regexCache := recache.NewCache(true)

	log.Debug("Setup graph.")

	// When we have resolved input ids, we can solve input changes.
	// Note: We do not want this to happen but really only if some `inputPathChanges`
	//       are given (can be []).
	resolveInputs := o.inputPathChanges != nil
	allNodes, allInputs, allComps, err := constructNodes(
		components,
		o.targetSelection,
		rootDir,
		resolveInputs,
	)
	if err != nil {
		return nil, nil, err
	}

	g, err := newGraph(allNodes, o.targetSelection)
	if err != nil {
		return nil, nil, err
	}

	err = g.SolveExecutionOrder()
	if err != nil {
		return nil, nil, err
	}

	// Make all input path changes absolute.
	for i := range o.inputPathChanges {
		o.inputPathChanges[i] = fs.MakeAbsoluteTo(rootDir, o.inputPathChanges[i])
	}
	log.Debug("Changed paths.", "paths", o.inputPathChanges)
	err = g.SolveInputChanges(allInputs, allComps, &regexCache, o.inputPathChanges)
	if err != nil {
		return nil, nil, err
	}

	targets, prios = g.NodesToPriorityList()

	return targets, prios, nil
}

func constructNodes(
	components []*component.Component,
	targetSelection *TargetSelection,
	rootDir string,
	resolveInputs bool,
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
				debug.Assertf(
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

	if resolveInputs {
		// Resolve inputs over all nodes on the graph.
		for _, n := range allNodes {
			e := resolveInputIDs(n, allInputs, allComps)
			if e != nil {
				return nil, nil, nil, e
			}
		}
	}

	return allNodes, allInputs, allComps, nil
}

func (graph *graph) recomputeSubgraph(selection *TargetSelection) error {
	if selection == nil {
		return nil
	} else if selection.Len() == 0 {
		return errors.New("graph selection must contain elements if not nil")
	}

	log.Debug("Recompute selection subgraph.")

	graph.execLeafNodesSel = &[]*TargetNode{}
	graph.execRootNodesSel = &[]*TargetNode{}

	for _, n := range graph.nodes {
		if selection.Exists(n.Target.ID) {
			selection.Remove(n.Target.ID)
			*graph.execLeafNodesSel = append(*graph.execLeafNodesSel, n)
		}
	}

	if selection.Len() != 0 {
		return errors.New(
			"not all target ids from selection were found in all "+
				"loaded components: remaining target ids: '%v'",
			selection,
		)
	}

	nodesSelection := set.NewUnorderedWithCap[target.ID](len(graph.nodes))

	execRootNodesSelection := filterNodesBFS(
		*graph.execLeafNodesSel,
		func(n *TargetNode) bool {
			nodesSelection.Insert(n.Target.ID)

			return len(n.Backward) == 0
		},
		backwardDir,
	)
	graph.execRootNodesSel = &execRootNodesSelection
	graph.nodesSel = &nodesSelection

	debug.Assert(
		len(*graph.execLeafNodesSel) == 0 || len(*graph.execRootNodesSel) != 0,
		"we found no root nodes from non-empty selection (this is a bug)",
	)

	log.Debug("Subgraph nodes.",
		"nodes", graph.nodesSel,
		"count", graph.nodesSel.Len())

	return nil
}

func newGraph(
	nodes TargetNodeMap,
	selection *TargetSelection) (graph, error) {
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
			continue
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

// resolveInputIDs resolves all `self|.*::XXX` in input ids in `.Inputs` and updates `allInputs`.
func resolveInputIDs(
	node *TargetNode,
	allInputs map[input.ID]*input.Config,
	allComps map[string]*component.Component,
) error {
	log.Tracef("Resolve input for node '%v'.", node.Target.ID)

	for idx, inputID := range node.Target.Inputs {
		log.Tracef("Resolve input id '%v'.", inputID)

		// Mangle `self` (referring to whole comp.)
		if inputID == "self" {
			inputID = input.DefineIDComp(node.Comp.Name())
			node.Target.Inputs[idx] = inputID
		} else if trimmedID, found := strings.CutPrefix(string(inputID), "self::"); found {
			// Mangle `self::` into own components input id.
			inputID = input.DefineID(node.Comp.Name(), trimmedID)
			node.Target.Inputs[idx] = inputID
		}

		if inputID.IsComponent() {
			if _, exists := allComps[string(inputID)]; !exists {
				return errors.New(
					"input id '%v' referring to a component in target '%v' on component '%v' is "+
						"not found on all found components\n"+
						"  -> working directory (or '-C') might be at the wrong place (use the top-level to check)",
					inputID,
					node.Target.ID,
					node.Config.Name,
				)
			}
		} else {
			if _, exists := allInputs[inputID]; !exists {
				return errors.New(
					"input id '%v' in target '%v' on component '%v' is "+
						"not found on all found components\n"+
						"  -> working directory (or '-C') might be at the wrong place (use the top-level to check)",
					inputID,
					node.Target.ID,
					node.Config.Name,
				)
			}
		}
	}

	log.Tracef("Resolved input ids '%v'.", node.Target.Inputs)

	return nil
}

// resolveTargetIDs resolves all `self::XXX` target ids in `.Dependencies`.
func resolveTargetIDs(node *TargetNode) {
	log.Debug("Resolve target ids.")
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
	log.Debug("Connect nodes.")

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
					"not be found by the loaded components\n"+
					"  -> working directory (or '-C') might be at the wrong place (use the top-level to check)",
				id)
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
					"dependency target id '%s' defined on target '%s' does not exist\n"+
						"  -> working directory (or '-C') might be at the wrong place (use the top-level to check)",
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

		log.Tracef("Start DFS cycle check: %v", root.Target.ID)
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
			log.Tracef("Current DFS path:\n%v", formatPath(&pathStack))

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
// visited nodes from a backwards traversal starting from all
// **changed** nodes in the **original selection**.
func (graph *graph) SolveInputChanges(
	inputs map[input.ID]*input.Config,
	comps map[string]*component.Component,
	regexCache *recache.Cache,
	paths []string,
) error {
	log.Debug("Solve input changes and recompute selection subgraph.")

	var changedInSelection TargetSelection
	inputChanges := make(map[input.ID]InputChanges, len(inputs))

	// Run the graph upwards (in execution order) and
	// propagate input changes and determine if
	// target is changed or not.
	for _, root := range *graph.execRootNodesSel {
		log.Tracef("Start from node: '%v'", root.Target.ID)
		dfsStack := stack.NewStack[*TargetNode]()
		dfsStack.Push(root)

		for dfsStack.Len() != 0 {
			log.Tracef("Current DFS stack:\n%v", formatStack(&dfsStack, false))
			n := dfsStack.Pop()

			if !graph.inSelection(n) {
				log.Trace("Skipped node not in selection.")

				continue
			}

			currIn := &n.Inputs

			switch {
			case paths == nil:
				log.Tracef("No paths given -> setting target id '%v' to changed.", n.Target.ID)
				currIn.Changed = true

			case !currIn.IsChanged():
				log.Tracef("Detect change for current node '%v'.", n.Target.ID)
				log.Tracef("Target inputs: '%v'", n.Target.Inputs)
				debug.Assert(paths != nil,
					"when we compute own changes, "+
						"we need a paths set and resolved inputs")

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
				log.Tracef("Current target '%s' already changed.", n.Target.ID)
				// we are skipped
			}

			log.Debug(
				"Changes for target id.",
				"id",
				n.Target.ID.String(),
				"changed",
				currIn.Changed,
				"changedByDeps",
				currIn.ChangedByDependency,
			)

			if currIn.IsChanged() {
				changedInSelection.Insert(n.Target.ID)
			}

			// Propagate to children, by merging the input changes.
			for _, c := range n.Forward {
				c.Inputs.Propagate(&n.Inputs)
			}

			// Go to next nodes.
			dfsStack.Push(n.Forward...)
		}
	}

	return graph.recomputeSubgraph(&changedInSelection)
}

// NodesToPriorityList converts the priorities on the nodes to an
// ordered list in descending order and returns all targets.
// It traverses the graph backwards starting from the lead nodes of the selection subgraph.
func (graph *graph) NodesToPriorityList() (nodes TargetNodeMap, result Priorities) {
	nodes = make(TargetNodeMap)
	prios := make(map[int]PrioritySet)

	log.Debug("Construct priority set.")
	visited := set.NewUnorderedWithCap[target.ID](len(graph.nodes))

	sel := func() iter.Seq[target.ID] {
		if graph.nodesSel != nil {
			return graph.nodesSel.Keys()
		}

		return maps.Keys(graph.nodes)
	}

	for id := range sel() {
		n, ok := graph.nodes[id]
		if !ok {
			log.Panicf("Node '%v' should be in the graph!", n.Target.ID)
		}

		if visited.Exists(n.Target.ID) {
			continue
		}

		debug.Assertf(graph.inSelection(n),
			"Node '%s' should be in selection at that point.", n.Target.ID)

		log.Tracef("Adding node '%v' with prio '%v'.", n.Target.ID, n.Priority)

		// Add the node to the result.
		prio := prios[n.Priority]
		prios[n.Priority] = PrioritySet{
			Priority: n.Priority,
			Nodes:    append(prio.Nodes, n),
		}
		nodes[n.Target.ID] = n

		visited.Insert(n.Target.ID)
	}

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
	log.Tracef("Check for changes in dir '%v'", rootDir)
	for i := range paths {
		debug.Assertf(path.IsAbs(paths[i]), "input path '%s' must be absolute", paths[i])
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
		// If no input ids, just add its own component, which is the default.
		inputID := input.DefineIDComp(node.Comp.Name())
		node.Target.Inputs = append(node.Target.Inputs, inputID)
	}

	// Check all input rules.
	for _, inputID := range node.Target.Inputs {
		log.Tracef("Check against input set '%v'.", inputID)

		inputCh := inputChanges[inputID]
		if inputCh.Changed {
			// Input changeset already determined.
			log.Tracef("Input changeset already changed.")

			return inputCh.Changed, inputCh.Changes, nil
		}

		// Not determined yet, lets check.
		// If the input id refers to a components, check against the root directory.
		if inputID.IsComponent() {
			comp := comps[string(inputID)]
			debug.Assert(comp != nil, "component referred is not existing")
			changed, changes := determineChangedPathsDefault(comp.Root(), paths)
			if changed {
				inputChanges[inputID] = InputChanges{Changed: changed, Changes: changes}

				// Lazy, do not check other input sets.
				return changed, changes, nil
			}

			continue
		}

		// otherwise check the input set
		log.Tracef("Check for changes for input id '%v'", inputID)
		input, exists := inputs[inputID]
		if !exists {
			return false, nil,
				errors.New("input id '%s' does not exist (programming error?)", inputID)
		}

		var relativePaths []string
		for i := range paths {
			if rel, matches := input.TrimOfBaseDir(paths[i]); matches {
				relativePaths = append(relativePaths, rel)
			}
		}

		if len(relativePaths) == 0 {
			// no paths which match this input set.
			log.Tracef("Relative paths do not match.")

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
			changed := includeRegexes.Match(p) && !excludeRegexes.Match(p)
			if changed {
				log.Tracef("Matched path '%v'.", p)
				inputCh = InputChanges{changed, []string{p}}
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

// Apply applies all options.
func (c *opts) Apply(options ...ExecOption) error {
	for _, f := range options {
		if err := f(c); err != nil {
			return err
		}
	}

	return nil
}

// WithTargetsByStageFromComponents turns components with optional stage into a target selection.
func WithTargetsByStageFromComponents(
	comps []*component.Component,
	stageFilter stage.Stage,
) ExecOption {
	return func(o *opts) error {
		if len(comps) > o.nodeCount {
			return errors.New("selection is longer than all nodes ('%v' > '%v')",
				len(comps), o.nodeCount,
			)
		}

		for _, comp := range comps {
			c := comp.Config()
			for _, tgt := range c.Targets {
				if stageFilter != "" && tgt.Stage != stageFilter {
					log.Debugf("Component '%v' does not matches stage.", tgt.ID)

					continue
				}

				log.Debugf("Component '%v' matches stage.", tgt.ID)

				if o.targetSelection == nil {
					s := set.NewUnordered[target.ID]()
					o.targetSelection = &s
				}

				debug.Assert(
					!o.targetSelection.Exists(tgt.ID),
					"target id '%v' should not exists",
					tgt.ID,
				)
				o.targetSelection.Insert(tgt.ID)
			}
		}

		return nil
	}
}

// WithTargetSelection sets a target selection directly.
func WithTargetSelection(sel *TargetSelection) ExecOption {
	return func(o *opts) error {
		o.targetSelection = sel

		return nil
	}
}

// WithTargetSelectionAdd adds target ids to the selection.
func WithTargetSelectionAdd(ids ...target.ID) ExecOption {
	return func(o *opts) error {
		if len(ids) != 0 && o.targetSelection == nil {
			s := set.NewUnorderedWithCap[target.ID](len(ids))
			o.targetSelection = &s
		}

		for i := range ids {
			o.targetSelection.Insert(ids[i])
		}

		return nil
	}
}

// WithInputChanges set the input path changes to be considered.
func WithInputChanges(inputPathChanges []string) ExecOption {
	return func(o *opts) error {
		o.inputPathChanges = inputPathChanges

		return nil
	}
}
