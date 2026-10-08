package dag

import (
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

// context implements the `runner.IContext` interface.
type context struct {
	gitx      git.Context
	comp      *component.Component
	targetID  target.ID
	toolchain string
	stepIdx   step.Index
	log       log.ILog
}

// Interface implementation guard.
var _ runner.IContext = (*context)(nil)

// Root implements [runner.IContext].
func (c *context) Root() string {
	return c.gitx.Cwd()
}

// Log implements [runner.IContext].
func (c *context) Log() log.ILog {
	return c.log
}

// Component implements [runner.IContext].
func (c *context) Component() *component.Component {
	return c.comp
}

// Target implements [runner.IContext].
func (c *context) Target() target.ID {
	return c.targetID
}

// Step implements [runner.IContext].
func (c *context) Step() step.Index {
	return c.stepIdx
}

// Toolchain implements [runner.IContext].
func (c *context) Toolchain() string {
	return c.toolchain
}

// Git implements [runner.IContext].
func (c *context) Git() git.Context {
	return c.gitx
}
