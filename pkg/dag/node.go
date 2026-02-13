package dag

import (
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
)

type (
	TargetNode struct {
		// The target to which this Node belongs.
		Target *target.Config

		// Where this target belongs to.
		Config *component.Config
		Comp   *component.Component

		// The execution priority.
		// The higher the earlier it should be executed.
		Priority int

		// All child nodes which depend on this node.
		// Forward in execution direction.
		Forward []*TargetNode

		// All nodes on which this node depends.
		// Backward in execution direction.
		Backward []*TargetNode

		// Tracking inputs on this node.
		Inputs TargetNodeChanges

		// Tracking execution.
		Execution TargetExecStatus
	}

	TargetExecStatus struct {
		// The target execution status.
		Status ExecStatus

		// Marking the target to not run and skip.
		Cancel bool

		// All runner statuses for the steps.
		Runners RunnerStatuses
	}

	TargetNodeChanges struct {
		// Flag if the inputs for this target have changed.
		Changed bool

		// ChangedByDependency denotes if the target has changed due to a dependency.
		// If so, then detection of own changed `Paths` are skipped!
		// Asserts: Can only be `true` if `Changed` is true and also then
		// `Paths` will be `nil` because we skip own detection.
		ChangedByDependency bool

		// Changed paths for this component.
		Paths []string

		// Changed paths by all parents.
		AccumulatedPaths []string
	}
)

// IsChanged returns the overall status if this node is changed.
func (i *TargetNodeChanges) IsChanged() bool {
	return i.Changed || i.ChangedByDependency
}

// Propagate propagates change state from `other` to `i`.
func (i *TargetNodeChanges) Propagate(other *TargetNodeChanges) {
	i.ChangedByDependency = i.ChangedByDependency || other.Changed
	i.AccumulatedPaths = append(i.AccumulatedPaths, other.AccumulatedPaths...)
}

// All returns all changes, accumulated ones from depend targets
// as well own changed paths.
func (i *TargetNodeChanges) All() []string {
	res := make([]string, 0, len(i.Paths)+len(i.AccumulatedPaths))
	res = append(res, i.Paths...)
	res = append(res, i.AccumulatedPaths...)

	return res
}

// AddRunnerStatus adds a runner status.
func (e *TargetExecStatus) AddRunnerStatus() *RunnerStatus {
	s := &RunnerStatus{}
	e.Runners = append(e.Runners, s)

	return s
}

// PropagateExecStatus propagates the execution status forward.
func (n *TargetNode) PropagateExecStatus() {
	for _, f := range n.Forward {
		if n.Execution.Status != ExecStatusSuccess {
			f.Execution.Cancel = true
		}
	}
}
