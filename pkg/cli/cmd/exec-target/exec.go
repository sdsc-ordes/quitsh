package exectarget

import (
	"github.com/sdsc-ordes/quitsh/pkg/cli"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/common/set"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/dag"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"

	"github.com/spf13/cobra"
)

type execTargetArgs struct {
	TargetIDs []string
}

func AddCmd(
	cli cli.ICLI,
	parent *cobra.Command,
) *cobra.Command {
	var args execTargetArgs
	execCmd := &cobra.Command{
		Use:          "exec-target [target-ids...]",
		Short:        "Execute a specific target on a component.",
		SilenceUsage: true,
		RunE: func(_cmd *cobra.Command, targs []string) error {
			args.TargetIDs = targs

			return runExec(cli, &args)
		},
	}

	_ = execCmd.MarkFlagRequired("component-dir")

	parent.AddCommand(execCmd)

	return execCmd
}

func runExec(cli cli.ICLI, args *execTargetArgs) error {
	log.Info("Executing target ...", "target-ids", args.TargetIDs)

	_, all, rootDir, err := cli.FindComponents(
		&general.ComponentArgs{ComponentPatterns: []string{"*"}},
	)
	if err != nil {
		return err
	}

	selection := set.NewUnorderedWithCap[target.ID](len(args.TargetIDs))
	for i := range args.TargetIDs {
		selection.Insert(target.ID(args.TargetIDs[i]))
	}

	targets, prios, err := dag.DefineExecutionOrder(all, &selection, nil, rootDir)
	if err != nil {
		return err
	}

	var dispatcher toolchain.IDispatcher
	if !cli.RootArgs().SkipToolchainDispatch {
		dispatcher = cli.ToolchainDispatcher()
	}

	if cli.RootArgs().Parallel {
		return dag.ExecuteDAGParallel(
			targets,
			cli.RunnerFactory(),
			dispatcher,
			cli.Config(),
			rootDir)
	} else {
		return dag.ExecuteDAG(
			prios,
			cli.RunnerFactory(),
			dispatcher,
			cli.Config(),
			rootDir,
		)
	}
}
