package dag

import (
	"fmt"

	taskflow "github.com/noneback/go-taskflow"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"
)

const MaxCoroutineConcurrency = 10000

// ExecuteConcurrent executes the DAG concurrent.
// TODO: Refactor to better length and less nested.
//
//nolint:gocognit,funlen
func ExecuteConcurrent(
	targetNodes TargetNodeMap,
	runnerFactory factory.IFactory,
	toolchainDispatcher toolchain.IDispatcher,
	config config.IConfig,
	rootDir string,
	opts ...ExecuteOption,
) error {
	opt := execOption{}
	if e := opt.Apply(opts...); e != nil {
		return e
	}

	executor := taskflow.NewExecutor(MaxCoroutineConcurrency)
	tf := taskflow.NewTaskFlow("DAG")

	var buildError error
	var summary Summary

	tasks := make(map[target.ID]*taskflow.Task, 0)
	for _, node := range targetNodes {
		tgtTask := tf.NewSubflow(node.Target.ID.String(),
			func(sf *taskflow.Subflow) {
				stepTasks := []*taskflow.Task{}

				for stepIdx := range node.Target.Steps {
					stepTask := sf.NewSubflow(
						fmt.Sprintf("%v::step-%v", node.Target.ID, stepIdx),

						func(sf *taskflow.Subflow) {
							e := addRunnerTasks(
								config, toolchainDispatcher, runnerFactory, rootDir,
								&summary,
								sf, node,
								&node.Target.Steps[stepIdx],
								stepIdx,
								&opt)

							buildError = errors.Combine(buildError, e)
						},
					)
					stepTasks = append(stepTasks, stepTask)
				}

				// Link all steps together.
				for i := 1; i < len(stepTasks); i++ {
					stepTasks[i].Succeed(stepTasks[i-1])
				}
			})

		tasks[node.Target.ID] = tgtTask
	}

	// Link together all target tasks.
	for id, task := range tasks {
		t := targetNodes[id]

		for _, b := range t.Backward {
			t, ok := tasks[b.Target.ID]
			if !ok {
				log.Panic(
					"Could not find task for target '%v'. " +
						"Something is wrong with the DAG!")
			}

			task.Succeed(t)
		}
	}

	if buildError != nil {
		return errors.AddContext(buildError,
			"Build errors while constructing the task flowk.")
	}

	executor.Run(tf).Wait()

	summary.statuses.Log()

	return summary.allErrors
}

func addRunnerTasks(
	config config.IConfig,
	toolchainDispatcher toolchain.IDispatcher,
	runnerFactory factory.IFactory,
	rootDir string,
	summary *Summary,
	sf *taskflow.Subflow,
	node *TargetNode,
	step *step.Config,
	stepIdx int,
	opt *execOption,
) error {
	var runners []factory.RunnerInstance
	var e error

	if !step.Include.TagExpr.Matches(opt.Tags) {
		log.Debugf(
			"Target: '%v' -> step idx: '%v' excluded: expr '%v' "+
				"does not match for tags '%q'",
			node.Target.ID, stepIdx,
			step.Include.TagExpr.String(), opt.Tags)

		return nil
	}

	if step.RunnerID != "" {
		runners, e = runnerFactory.CreateByID(
			step.RunnerID, step.Toolchain, step.ConfigRaw)
	} else if step.Runner != "" {
		runners, e = runnerFactory.CreateByKey(
			runner.NewRegisterKey(node.Target.Stage, step.Runner),
			step.Toolchain,
			step.ConfigRaw,
		)
	}

	if e != nil {
		return errors.AddContext(e,
			"could not instantiate runner for target '%v'", node.Target.ID)
	}

	runnerTasks := []*taskflow.Task{}
	for runnerIdx, r := range runners {
		n := fmt.Sprintf("%v::step-%v::%v", node.Target.ID, step.Index, runnerIdx)

		logger := log.NewLogger(node.Target.ID.String())

		runnerTask := sf.NewTask(
			n,
			func() {
				var stat ExecStatus = ExecStatusNotRun
				var err error
				defer func() {
					if r := recover(); r != nil {
						err = errors.Combine(err, errors.New("panic in runner task '%v': %v", n, r))
					}

					if err == nil {
						stat = ExecStatusSuccess
					} else {
						stat = ExecStatusFailed
					}

					summary.AddStatus(
						RunnerStatus{
							stat,
							err,
							node.Comp.Name(),
							node.Target.ID,
							step.Index,
							r.RunnerID,
						},
					)
				}()

				err = ExecuteRunner(
					logger,
					node.Comp,
					node.Target.ID,
					step.Index,
					runnerIdx,
					r.Runner,
					r.Toolchain,
					toolchainDispatcher,
					config,
					rootDir,
				)
			})

		runnerTasks = append(runnerTasks, runnerTask)
	}

	// Link all runners on this step together.
	for i := 1; i < len(runnerTasks); i++ {
		runnerTasks[i].Succeed(runnerTasks[i-1])
	}

	return summary.allErrors
}
