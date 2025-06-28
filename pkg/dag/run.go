package dag

import (
	"fmt"
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/component/target"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"
)

type RunnerData struct {
	comp      *component.Component
	targetID  target.ID
	step      *step.Config
	runnerIdx int
	inst      factory.RunnerInstance
}

// ExecuteDAG executes the DAG.
// If not dispatcher is given, the toolchain dispatch is not done.
func ExecuteDAG(
	prios Priorities,
	runnerFactory factory.IFactory,
	toolchainDispatcher toolchain.IDispatcher,
	config config.IConfig,
	rootDir string,
) error {
	currCwd, err := os.Getwd()
	log.PanicE(err, "Cannot get working dir.")
	defer func() { _ = os.Chdir(currCwd) }()

	allRunners := []RunnerData{}

	addRunners := func(node *TargetNode, step *step.Config) {
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
			e = errors.AddContext(e, "could not instantiate runner for target '%v'", node.Target.ID)
			err = errors.Combine(err, e)

			return
		}

		for runnerIdx, r := range runners {
			allRunners = append(allRunners,
				RunnerData{
					comp:      node.Comp,
					targetID:  node.Target.ID,
					step:      step,
					runnerIdx: runnerIdx,
					inst:      r,
				})
		}
	}

	for _, prio := range prios {
		for _, node := range prio.Nodes {
			for _, step := range node.Target.Steps {
				addRunners(node, &step)
			}
		}
	}

	if err != nil {
		return errors.AddContext(err, "failed to assemble all runners")
	}

	return executeRunners(
		allRunners,
		toolchainDispatcher,
		config,
		rootDir,
	)
}

func executeRunners(
	allRunners []RunnerData,
	toolchainDispatcher toolchain.IDispatcher,
	config config.IConfig,
	rootDir string,
) error {
	log.Info("Collected runners.", "count", len(allRunners))

	var summary RunnerStatuses
	var err error

	for _, rD := range allRunners {
		failed := false

		log.Info("Starting runner.", "runner", rD.inst.RunnerID, "target", rD.targetID)

		e := ExecuteRunner(
			rD.comp,
			rD.targetID,
			rD.step.Index,
			rD.runnerIdx,
			rD.inst.Runner,
			rD.inst.Toolchain,
			toolchainDispatcher,
			config,
			rootDir,
			false,
		)

		if e != nil {
			e = errors.AddContext(e,
				"Runner '%v' for target '%v' failed.",
				rD.inst.RunnerID,
				rD.targetID)
			failed = true
			err = errors.Combine(err, e)
		}

		summary = append(
			summary,
			RunnerStatus{
				failed,
				rD.comp.Root(),
				rD.targetID,
				rD.step.Index,
				rD.inst.RunnerID,
			},
		)
	}

	summary.Log()

	return err
}

func ExecuteRunner(
	comp *component.Component,
	targetID target.ID,
	stepIdx step.Index,
	runnerIdx int,
	runner runner.IRunner,
	toolchainName string,
	toolchainDispatcher toolchain.IDispatcher,
	config config.IConfig,
	rootDir string,
	addPrefix bool,
) error {
	// When the toolchain is 'none', none is needed.
	skipDispatch := toolchainDispatcher == nil
	haveToolchain := nix.HaveToolchain(toolchainName)
	noDispatch := skipDispatch || haveToolchain

	log.Info("Start runner.", "step", stepIdx, "runner", runner.ID())
	log.Info(
		"Toolchain status.",
		"Name", toolchainName,
		"SkipDispatch",
		skipDispatch,
		"HaveToolchain",
		haveToolchain,
	)

	if skipDispatch {
		if !haveToolchain {
			log.Panic(
				"Something is wrong: \n"+
					"We dispatch over the toolchain but the toolchain is not detected.",
				"toolchain", toolchainName,
			)
		}
	}

	var logPrefix string

	if addPrefix {
		logPrefix = fmt.Sprintf("[%s]", targetID.String())
	}

	if noDispatch { //nolint: nestif,nolintlint
		// Change to repo root and run the runner.
		err := os.Chdir(rootDir)
		if err != nil {
			return err
		}

		ctx := context{
			gitx:      git.NewCtx(rootDir),
			comp:      comp,
			targetID:  targetID,
			toolchain: toolchainName,
			stepIdx:   stepIdx,
			log:       log.NewLogger(logPrefix),
		}
		err = runner.Run(&ctx)

		if err != nil {
			return err
		}

		log.Info("Runner successful.", "runner", runner.ID(), "target", targetID)
	} else {
		if toolchainDispatcher == nil {
			log.Panic("Something is wrong: \n"+
				"We dispatch over the toolchain but no dispatcher is given.",
				"toolchain", toolchainName)
		}

		dArgs := toolchain.DispatchArgs{
			ComponentDir: comp.Root(),
			TargetID:     targetID,
			StepIndex:    stepIdx,
			RunnerIndex:  runnerIdx,
			RunnerID:     runner.ID(),
			Toolchain:    toolchainName}
		err := toolchainDispatcher.Run(rootDir, &dArgs, config)

		if err != nil {
			log.Info("Toolchain dispatch failed.", "runner", runner.ID(), "target", targetID)

			return err
		}

		log.Info("Toolchain dispatch successful.", "runner", runner.ID(), "target", targetID)
	}

	return nil
}

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
