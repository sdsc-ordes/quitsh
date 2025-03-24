package dag

import (
	"fmt"
	"runtime"
	"sync"

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

func ExecuteDAGParallel(
	targetNodes TargetNodeMap,
	runnerFactory factory.IFactory,
	toolchainDispatcher toolchain.IDispatcher,
	config config.IConfig,
	rootDir string,
) (allErrors error) {
	executor := taskflow.NewExecutor(uint(runtime.NumCPU()-1) * 10000) //nolint:gosec,mnd
	tf := taskflow.NewTaskFlow("DAG")

	var buildError error
	var lock sync.Mutex

	addRunnerTasks := func(sf *taskflow.Subflow, node *TargetNode, step *step.Config) {
		var runners []factory.RunnerInstance
		var e error

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
			buildError = errors.Combine(buildError,
				errors.AddContext(e,
					"could not instantiate runner for target '%v'", node.Target.ID))

			return
		}

		runnerTasks := []*taskflow.Task{}
		for runnerIdx, r := range runners {
			n := fmt.Sprintf("%v::step-%v::%v", node.Target.ID, step.Index, runnerIdx)
			runnerTask := sf.NewTask(
				n,
				func() {
					var err error
					defer func() {
						var p error
						if r := recover(); r != nil {
							p = errors.New("panic in runner task '%v': %v", n, r)
						}

						lock.Lock()
						defer lock.Unlock()
						allErrors = errors.Combine(allErrors, err, p)
					}()

					err = ExecuteRunner(
						node.Comp,
						node.Target.ID,
						step.Index,
						runnerIdx,
						r.Runner,
						r.Toolchain,
						toolchainDispatcher,
						config,
						rootDir)
				})

			runnerTasks = append(runnerTasks, runnerTask)
		}

		// Link all runners on this step together.
		for i := 1; i < len(runnerTasks); i++ {
			runnerTasks[i].Succeed(runnerTasks[i-1])
		}
	}

	tasks := make(map[target.ID]*taskflow.Task, 0)
	for _, node := range targetNodes {
		tgtTask := tf.NewSubflow(node.Target.ID.String(),
			func(sf *taskflow.Subflow) {
				stepTasks := []*taskflow.Task{}
				for i := range node.Target.Steps {
					stepTask := sf.NewSubflow(fmt.Sprintf("%v::step-%v", node.Target.ID, i),
						func(sf *taskflow.Subflow) {
							addRunnerTasks(sf, node, &node.Target.Steps[i])
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
				log.Panic("Could not find task for target '%v'. Something is wrong with the DAG!")
			}

			task.Succeed(t)
		}
	}

	if buildError != nil {
		return buildError
	}

	executor.Run(tf).Wait()

	return allErrors
}
