package gorunner

import (
	"path"
	"quitsh-cli/pkg/runner/config"
	"quitsh-cli/pkg/setup"
	"slices"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	gox "github.com/sdsc-ordes/quitsh/pkg/exec/go"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

const GoLintRunnerID = "cli::lint-go"

type GoLintRunner struct {
	runnerConfig *RunnerConfigLint
	settings     *config.LintSettings
}

type RunnerConfigLint struct {
}

func UnmarshalLintConfig(_raw step.AuxConfigRaw) (step.AuxConfig, error) {
	return &RunnerConfigLint{}, nil
}

func NewGoLintRunner(config any, settings *config.LintSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &GoLintRunner{
		runnerConfig: common.Cast[*RunnerConfigLint](config),
		settings:     settings,
	}, nil
}

func getFlags(rootDir string) (flags []string) {
	flags = append(flags,
		"--max-issues-per-linter", "0",
		"--max-same-issues", "0",
		"--timeout", "20m0s",
		"--verbose",
		"--config",
		path.Join(rootDir, ".golangci.yaml"))

	return
}

func (r *GoLintRunner) ID() runner.RegisterID {
	return GoLintRunnerID
}

func (r *GoLintRunner) Run(ctx runner.IContext) error {
	comp := ctx.Component()

	err := runGoModTidy(ctx.Log(), comp)

	e := runGoLangCILint(ctx.Log(), comp, ctx.Root())
	err = errors.Combine(e, err)

	return err
}

func runGoModTidy(log log.ILog, comp *component.Component) error {
	log.Info("Starting `no-go-mod-tidy-changes`.", "component", comp.Config().Name)

	goctx := gox.NewCtxBuilder().
		Cwd(comp.Root()).
		Build()

	err := goctx.Check("mod", "tidy")
	if err != nil {
		return err
	}

	gitx := git.NewCtx(comp.Root())
	files, err := gitx.Changes(".", false)
	if err != nil {
		return err
	}

	if slices.Contains(files, "go.mod") || slices.Contains(files, "go.sum") {
		log.Error("Detected 'go.mod' changes.")

		return errors.New(
			"Go mod file in '%v' is not correct and has changed due to `go mod tidy`.",
			comp.Root(),
		)
	}

	return nil
}

func runGoLangCILint(log log.ILog, comp *component.Component, rootDir string) error {
	log.Info("Starting `golangcilint` for component.", "component", comp.Config().Name)

	err := setup.LinkConfigFiles(rootDir)
	if err != nil {
		return err
	}

	lintctx := exec.NewCmdCtxBuilder().
		BaseCmd("golangci-lint").
		Cwd(comp.Root()).
		ExitCodeHandler(
			func(err *exec.CmdError) error {
				switch {
				case err == nil:
					return nil
				case err.ExitCode() == 1:
					log.Error("Go lint errors detected, see output above.")

					return errors.New("golangci-lint lint errors")
				default:
					return err
				}
			}).
		Build()

	flags := getFlags(rootDir)
	cmd := append([]string{"run"}, flags...)
	cmd = append(cmd, "./...")

	return lintctx.Check(cmd...)
}
