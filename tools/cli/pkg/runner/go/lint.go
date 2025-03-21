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
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

const GoLintRunnerID = "custodian::lint-go"

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
		"--print-resources-usage",
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

	e = runNoWrongIncludes(ctx.Log(), comp)
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

func runNoWrongIncludes(log log.ILog, comp *component.Component) error {
	log.Info("Starting `no-wrong-includes`.", "component", comp.Config().Name)

	var noIncAs []string

	grep := exec.NewCmdCtxBuilder().
		BaseCmd("grep").
		BaseArgs(
			"-r",
			"-H",
			"--perl-regexp",
			"--exclude-dir", fs.OutputDir,
			"--exclude-dir='.git'",
			"--include", "*.go", "--include", "go.mod",
			"-I", // no binary fs.
		).

		// BaseArgs("--hidden", "-n", "--glob", "*.go", "--glob", "go.mod").
		ExitCodeHandler(func(e *exec.CmdError) error {
			switch {
			case e == nil:
				log.Error("Inacceptable includes detected, see above")

				return errors.New("inacceptable includes detected")
			case e.ExitCode() == 1:
				return nil
			default:
				return e
			}
		}).
		Build()

	// Strings are concatenated to
	// to circumvent detection in this file.
	switch comp.Config().Name {
	case "quitshv2":
		// Must not include anything from components, must be self-containing.
		noIncAs = append(noIncAs, "custodian"+"/components")
	case "lib-common":
		// Must not include anything else, must be self-containing.
		noIncAs = append(
			noIncAs,
			"custodian"+"/components/(?!lib-common)",
			"custodian"+"/tools/quitshv2",
		)
	default:
		// all other components must not include `quitshv2`
		// TODO: This test needs to become better to basically disallow
		// importing other components (maybe negative lookahead).
		noIncAs = append(
			noIncAs,
			"custodian"+"/tools/quitshv2",
		)
	}

	var err error
	for _, inc := range noIncAs {
		e := grep.Check(inc, comp.Root())
		err = errors.Combine(e, err)
	}

	return err
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
