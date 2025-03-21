package execrunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/dag"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"

	"github.com/spf13/cobra"
)

func AddCmd(
	cli cli.ICLI,
	parent *cobra.Command,
	args *toolchain.DispatchArgs,
) *cobra.Command {
	execCmd := &cobra.Command{
		Use:          "exec-runner",
		Short:        "Execute a specific runner on a component->target->step.",
		SilenceUsage: true,
		RunE: func(_cmd *cobra.Command, _args []string) error {
			return runExec(cli, args)
		},
	}

	execCmd.Flags().
		StringVarP(&args.ComponentDir,
			"component-dir", "c",
			".",
			"The component directory.",
		)

	execCmd.Flags().
		StringVarP((*string)(&args.TargetID),
			"target-id", "t",
			"",
			"The target id on this component.",
		)

	execCmd.Flags().
		IntVarP((*int)(&args.StepIndex),
			"step-index", "s",
			-1,
			"The step index to execute on the target.",
		)
	execCmd.Flags().
		IntVarP(&args.RunnerIndex,
			"runner-index", "r",
			-1,
			"The runner index in the array of returned runners on the step. "+
				"Normally you return one runner, then its '0'.",
		)

	parent.AddCommand(execCmd)

	return execCmd
}

func runExec(cli cli.ICLI, args *toolchain.DispatchArgs) error {
	log.Info("Executing runner ...", "args", args)

	err := args.Validate()
	if err != nil {
		return errors.AddContext(err, "input arguments are not correct")
	}

	comps, _, rootDir, err := cli.FindComponents(
		&general.ComponentArgs{ComponentDir: args.ComponentDir},
	)
	if err != nil {
		return err
	}

	if len(comps) != 1 {
		return errors.New("we should only find one component")
	}

	comp := comps[0]
	target := comp.Config().TargetByID(args.TargetID)
	if target == nil {
		return errors.New(
			"target id '%v' is not existing on component '%v'",
			args.TargetID,
			comp.Name(),
		)
	}

	if int(args.StepIndex) >= len(target.Steps) {
		return errors.New(
			"step index '%v' is out of bound in target id '%v'",
			args.StepIndex,
			args.TargetID,
		)
	}
	step := &target.Steps[args.StepIndex]

	var runners []factory.RunnerInstance
	if step.RunnerID != "" {
		if args.RunnerID != "" && args.RunnerID != step.RunnerID {
			return errors.New(
				"runner id given is '%v' but step defines '%v'",
				args.RunnerID,
				step.RunnerID,
			)
		}

		runners, err = cli.RunnerFactory().CreateByID(
			step.RunnerID, step.Toolchain, step.ConfigRaw)
	} else if step.Runner != "" {
		runners, err = cli.RunnerFactory().CreateByKey(
			runner.NewRegisterKey(target.Stage, step.Runner),
			step.Toolchain,
			step.ConfigRaw,
		)
	}

	if err != nil {
		return err
	}

	if args.RunnerIndex >= len(runners) {
		return errors.New(
			"runner index '%v' is out of bound for returned runner count in target id '%v', step index: '%v'",
			args.RunnerIndex,
			args.TargetID,
			args.StepIndex,
		)
	}
	runner := runners[args.RunnerIndex]

	if args.Toolchain != "" && args.Toolchain != runner.Toolchain {
		return errors.New(
			"toolchain given is '%v' but runner uses '%v'",
			args.RunnerID,
			step.RunnerID,
		)
	}

	var dispatcher toolchain.IDispatcher
	if !cli.RootArgs().SkipToolchainDispatch {
		dispatcher = cli.ToolchainDispatcher()
	}

	return dag.ExecuteRunner(
		comp,
		target.ID,
		step.Index,
		args.RunnerIndex,
		runner.Runner,
		runner.Toolchain,
		dispatcher,
		cli.Config(),
		rootDir,
	)
}
