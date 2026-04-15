package errorsfilter

import (
	"github.com/hashicorp/go-multierror"
	"github.com/sdsc-ordes/quitsh/pkg/common/stack"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// alreadyReportedErr is an wrapping error which can be dropped by `DropReported`.
type alreadyReportedErr struct {
	e error
}

func (e *alreadyReportedErr) Error() string {
	return e.e.Error()
}

func (e *alreadyReportedErr) Unwrap() error {
	return e.e
}

// WrapAsReported returns a new wrapped errors which indicates
// that it is already reported.
// Useful for bigger errors, which you want to propagate upwards but
// maybe ignore in logging with [FilterAlreadyReported].
func WrapAsReported(e error) error {
	return &alreadyReportedErr{e: e}
}

// FilterAlreadyReported filters `ErrAlreadyReported` from the chain.
// NOTE: This only filters on [multierror.Error] s since, `Unwrap` functionality does not let
// reconstruct the error. We are using [multierror.Error] extensively anyways.
//
//nolint:gocognit,nestif // TODO: later.
func FilterAlreadyReported(err error) error {
	if err == nil {
		return nil
	}

	type Node struct {
		parent *Node
		e      error

		multi   *multierror.Error
		new     error
		visited bool // Denotes if we visited the node on down traversal.
	}

	stack := stack.NewStack[*Node]()
	root := &Node{e: err}
	stack.Push(root)

	for stack.Len() != 0 {
		node := stack.Top()
		if node.visited {
			// Backtracking
			// Reconstruct errors.
			ignore := false
			if _, ok := node.e.(*alreadyReportedErr); ok { //nolint:errorlint // This is ok.
				ignore = true
			}

			if !ignore {
				// Create our new error from any children we have.
				if node.multi != nil {
					node.new = node.multi
				} else {
					node.new = node.e
				}

				// Add ourself to parent.
				if node.parent != nil {
					node.parent.multi.Errors = append(node.parent.multi.Errors, node.new)
				}
			} else {
				log.Debug("Dropping error 'alreadyReportedErr'.")
			}

			stack.Pop()

			continue
		}

		node.visited = true
		if e, ok := node.e.(*multierror.Error); ok { //nolint:errorlint // This is ok.
			node.multi = &multierror.Error{} // init a new multi list error
			for _, c := range e.Errors {
				stack.Push(&Node{e: c, parent: node})
			}
		}
	}

	return root.new
}
