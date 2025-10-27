package toolchain

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

// The dispatcher interface which will
// run the runner over a toolchain.
type IDispatcher interface {
	Run(
		repoDir string,
		dispatchArgs *DispatchArgs,
		config config.IConfig,
	) error
}

type DispatchArgs struct {
	ComponentDir string     `yaml:"componentDir" validate:"required"`
	TargetID     target.ID  `yaml:"targetID"     validate:"required"`
	StepIndex    step.Index `yaml:"stepIndex"    validate:"gte=0"`
	RunnerIndex  int        `yaml:"runnerIndex"  validate:"gte=0"`

	RunnerID  runner.RegisterID `yaml:"runnerID"  validate:""`
	Toolchain string            `yaml:"toolchain" validate:""`
}

// Validate validates the dispatch args.
func (c *DispatchArgs) Validate() error {
	return common.Validator().Struct(c)
}
