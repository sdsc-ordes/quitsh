package dag

import (
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
	"github.com/sdsc-ordes/quitsh/pkg/tags"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"
)

type (
	RunnerData struct {
		node      *TargetNode
		comp      *component.Component
		status    *RunnerStatus
		targetID  target.ID
		step      *step.Config
		runnerIdx int
		inst      factory.RunnerInstance
	}

	ExecArgs struct {
		Tags []string `yaml:"tags"`
	}

	ExecuteOption func(*execOption) error

	execOption struct {
		Tags []tags.Tag
	}
)

// Execute executes the DAG.
// If no dispatcher is given, the toolchain dispatch is not done.
func Execute(
	targets TargetNodeMap,
	prios Priorities,
	runnerFactory factory.IFactory,
	dispatcher toolchain.IDispatcher,
	config config.IConfig,
	rootDir string,
	parallel bool,
	opts ...ExecuteOption,
) error {
	if parallel {
		return ExecuteConcurrent(
			targets,
			runnerFactory,
			dispatcher,
			config,
			rootDir, opts...)
	} else {
		return ExecuteNormal(
			prios,
			runnerFactory,
			dispatcher,
			config,
			rootDir, opts...,
		)
	}
}

// ExecuteNormal executes the DAG non-concurrent.
// If no dispatcher is given, the toolchain dispatch is not done.
func ExecuteNormal(
	prios Priorities,
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

	currCwd, err := os.Getwd()
	log.PanicE(err, "Cannot get working dir.")
	defer func() { _ = os.Chdir(currCwd) }()

	allRunners := []RunnerData{}

	addRunners := func(node *TargetNode, step *step.Config, stepIdx int) {
		var runners []factory.RunnerInstance
		var e error

		if !step.Include.TagExpr.Matches(opt.Tags) {
			log.Warnf(
				"Target: '%v' -> Step: '%v' excluded: expr '%v' "+
					"does not match for tags '%q'",
				node.Target.ID, stepIdx,
				step.Include.TagExpr.String(), opt.Tags)

			return
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
			e = errors.AddContext(e,
				"could not instantiate runner for target '%v'",
				node.Target.ID)
			err = errors.Combine(err, e)

			return
		}

		for runnerIdx, r := range runners {
			status := node.Exec.AddRunnerStatus()

			allRunners = append(allRunners,
				RunnerData{
					node:      node,
					comp:      node.Comp,
					status:    status,
					targetID:  node.Target.ID,
					step:      step,
					runnerIdx: runnerIdx,
					inst:      r,
				})
		}
	}

	for _, prio := range prios {
		for _, node := range prio.Nodes {
			for stepIdx, step := range node.Target.Steps {
				addRunners(node, &step, stepIdx)
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

	var summary Summary

	for _, rD := range allRunners {
		var stat ExecStatus = ExecStatusNotRun

		log.Info("Starting runner.", "runner", rD.inst.RunnerID, "target", rD.targetID)

		e := ExecuteRunner(
			log.NewLogger(rD.targetID.String()),
			rD.comp,
			rD.targetID,
			rD.step.Index,
			rD.runnerIdx,
			rD.inst.Runner,
			rD.inst.Toolchain,
			toolchainDispatcher,
			config,
			rootDir,
		)

		if e != nil {
			e = errors.AddContext(e,
				"Runner '%v' for target '%v' failed.",
				rD.inst.RunnerID,
				rD.targetID)
			stat = ExecStatusFailed
		} else {
			stat = ExecStatusSuccess
		}

		*rD.status =
			RunnerStatus{
				stat,
				e,
				rD.comp.Root(),
				rD.targetID,
				rD.step.Index,
				rD.inst.RunnerID,
			}

		summary.AddStatus(rD.status)
	}

	summary.statuses.Log()

	return summary.allErrors
}

func ExecuteRunner(
	log log.ILog,
	comp *component.Component,
	targetID target.ID,
	stepIdx step.Index,
	runnerIdx int,
	runner runner.IRunner,
	toolchainName string,
	toolchainDispatcher toolchain.IDispatcher,
	config config.IConfig,
	rootDir string,
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
			log:       log,
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

// Apply applies option `opts` to `execOption`.
func (o *execOption) Apply(opts ...ExecuteOption) error {
	for i := range opts {
		e := opts[i](o)
		if e != nil {
			return errors.AddContext(e, "could not apply execute option")
		}
	}

	return nil
}

// WithTags adds executable tags [tags.Tag] to the executable options.
func WithTags(tag ...string) ExecuteOption {
	return func(o *execOption) error {
		for i := range tag {
			o.Tags = append(o.Tags, tags.NewTag(tag[i]))
		}

		return nil
	}
}
