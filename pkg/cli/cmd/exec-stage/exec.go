package execstage

import (
	"fmt"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/dag"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"

	"github.com/spf13/cobra"
)

type (
	Option func(*opts)

	opts struct {
		name   string
		modify func(cmd *cobra.Command)
	}
)

// AddCmdGeneral adds the general `exec-stage` command to execute all targets
// in given stages.
func AddCmdGeneral(
	cli cli.ICLI,
	parent *cobra.Command,
	execArgs *dag.ExecArgs,
) {
	var compArgs general.ComponentArgs
	cmd := &cobra.Command{
		Use:   "exec-stage [stage...]",
		Short: "Execute all targets in a stage.",
		RunE: func(_ *cobra.Command, stages []string) error {
			for _, s := range stages {
				e := ExecuteStage(cli, &compArgs, stage.Stage(s), execArgs)
				if e != nil {
					return e
				}
			}

			return nil
		},
	}
	general.AddFlagsExecArgs(cmd, execArgs)
	general.AddFlagsComponentArgs(cmd, &compArgs)

	parent.AddCommand(cmd)
}

// AddCmdAlias adds a generalized command similar to `AddCmdGeneral` but
// with name either from [WithName] or defaulted to `stage`.
// Its possible to modify the command with [WithModifications].
func AddCmdAlias(
	cli cli.ICLI,
	parent *cobra.Command,
	stage stage.Stage,
	execArgs *dag.ExecArgs,
	opt ...Option,
) {
	var o opts
	o.Apply(opt...)

	if o.name == "" {
		o.name = stage.String()
	}

	var compArgs general.ComponentArgs

	cmd := &cobra.Command{
		Use:   o.name,
		Short: fmt.Sprintf("Execute all targets in stage %v.", stage),
		RunE: func(_ *cobra.Command, _args []string) error {
			return ExecuteStage(cli, &compArgs, stage, execArgs)
		},
	}

	general.AddFlagsExecArgs(cmd, execArgs)
	general.AddFlagsComponentArgs(cmd, &compArgs)

	if o.modify != nil {
		o.modify(cmd)
	}

	parent.AddCommand(cmd)
}

// ExecuteStage executes all targets found with `compArgs` which belong to stage `stage`.
func ExecuteStage(
	cl cli.ICLI,
	compArgs *general.ComponentArgs,
	stage stage.Stage,
	execArgs *dag.ExecArgs,
) error {
	comps, all, rootDir, err := cl.FindComponents(compArgs)
	if err != nil {
		return err
	}

	targets, prios, err := dag.DefineExecutionOrder(
		all, rootDir,
		dag.WithTargetsByStageFromComponents(comps, stage),
	)
	if err != nil {
		return err
	} else if len(targets) == 0 {
		return errors.New("no targets selected")
	}

	var dispatcher toolchain.IDispatcher
	if !cl.RootArgs().SkipToolchainDispatch {
		dispatcher = cl.ToolchainDispatcher()
	}

	return dag.Execute(
		targets,
		prios,
		cl.RunnerFactory(),
		dispatcher,
		cl.Config(),
		rootDir,
		cl.RootArgs().Parallel,
		dag.WithTags(execArgs.Tags...),
	)
}

// Apply applies all options.
func (c *opts) Apply(options ...Option) {
	for _, f := range options {
		f(c)
	}
}

// WithName sets the command name to use.
func WithName(name string) Option {
	return func(o *opts) {
		o.name = name
	}
}

// WithModifications sets a modification function to change the command.
func WithModifications(mod func(cmd *cobra.Command)) Option {
	return func(o *opts) {
		o.modify = mod
	}
}
