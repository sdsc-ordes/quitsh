package runner

import (
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type IContext interface {
	// The root directory of the repository.
	Root() string

	// The Git context initialized at the root of the repository.
	Git() git.Context

	// The logger object.
	Log() log.ILog

	// On which component the runner executes.
	// Should not be able to change things in here.
	Component() *component.Component

	// On which target and step the runner executes.
	Target() target.ID

	// On which step the runner executes.
	Step() step.Index

	// The toolchain this runner runs in.
	Toolchain() string
}
