package dag

import (
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/log"
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

func (c *context) Root() string {
	return c.gitx.Cwd()
}

func (c *context) Log() log.ILog {
	return c.log
}

func (c *context) Component() *component.Component {
	return c.comp
}

func (c *context) Target() target.ID {
	return c.targetID
}

func (c *context) Step() step.Index {
	return c.stepIdx
}

func (c *context) Toolchain() string {
	return c.toolchain
}

func (c *context) Git() git.Context {
	return c.gitx
}
