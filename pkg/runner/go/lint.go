package gorunner

import (
	"bufio"
	"go/build/constraint"
	"os"
	"path"
	"slices"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	gox "github.com/sdsc-ordes/quitsh/pkg/exec/go"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const GoLintRunnerID = "custodian::lint-go"

type (
	GoLintRunner struct {
		runnerConfig *LintConfig
		settings     *config.LintSettings
	}
)

func NewGoLintRunner(config any, settings *config.LintSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &GoLintRunner{
		runnerConfig: common.Cast[*LintConfig](config),
		settings:     settings,
	}, nil
}

func getGoLangCILintFlags(configPath string, fix bool) (flags []string) {
	flags = append(flags,
		"--config", configPath,
		"--allow-parallel-runners",
		"--max-issues-per-linter", "0",
		"--max-same-issues", "0",
		"--timeout", "20m0s",
		"--verbose")

	if fix {
		flags = append(flags, "--fix")
	}

	return
}

// Interface implementation guard.
var _ runner.IRunner = (*GoLintRunner)(nil)

// ID implements [runner.IRunner].
func (r *GoLintRunner) ID() runner.RegisterID {
	return GoLintRunnerID
}

// Run implements [runner.IRunner].
func (r *GoLintRunner) Run(ctx runner.IContext) error {
	comp := ctx.Component()

	err := runGoModTidy(ctx.Log(), comp)

	e := runGoLangCILint(ctx.Log(), r.settings, &r.runnerConfig.GolangCILint, comp, ctx.Root())
	err = errors.Combine(e, err)

	if r.runnerConfig.CheckBuildConstraints.Enable {
		e = runTestBuildConstraints(ctx.Log(), comp, r.runnerConfig.CheckBuildConstraints.Rules)
		err = errors.Combine(err, e)
	}

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

func runTestBuildConstraints(
	log log.ILog,
	comp *component.Component,
	rules []BuiltConstraintRule) (err error) {
	for i := range rules {
		err = errors.Combine(err, runTestBuildConstraintsRule(log, comp, &rules[i]))
	}

	return err
}

func runTestBuildConstraintsRule(
	log log.ILog,
	comp *component.Component,
	rule *BuiltConstraintRule,
) (err error) {
	log.Info("Running build tags check on test files.")
	files, _, e := fs.FindFilesByPatterns(comp.Root(),
		rule.IncludePatterns, nil, fs.WithWalkDirFilterDefault(true))
	if e != nil {
		return errors.AddContext(err, "could not find Go test files")
	}

	const buildTags = "//go:build"
	log.Infof("Checking '%v' test files for build constraints.", len(files))
	if len(files) == 0 {
		return errors.New("matched '0' files in build constraint check")
	}

	for _, file := range files {
		f, e2 := os.Open(file)
		if e2 != nil {
			return e2
		}

		scanner := bufio.NewScanner(f)
		if ok := scanner.Scan(); !ok {
			err = errors.Combine(err,
				errors.New("File '%s' does not contain one line!", file))
		}

		firstLine := scanner.Text()

		expr, e2 := constraint.Parse(firstLine)
		if e2 != nil {
			err = errors.Combine(err, e2,
				errors.New("Go test file '%s' does not contain a necessary"+
					"build constraint starting with '%s' on first line", file, buildTags))
		}

		for i := range rule.Constraints {
			if expr.String() != rule.Constraints[i] {
				err = errors.Combine(err,
					errors.New("Go test file '%s' does not contain a necessary "+
						"build constraint '%s'", file, rule.Constraints[i]))
			}
		}
	}

	return err
}

func runGoLangCILint(
	log log.ILog,
	sett *config.LintSettings,
	config *GolangCILint,
	comp *component.Component,
	rootDir string,
) error {
	log.Info("Starting `golangcilint` for component.", "component", comp.Config().Name)

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

	configFile := path.Join(comp.Root(), config.Config)
	if !fs.Exists(configFile) {
		configFile = path.Join(rootDir, config.RootConfig)
	}

	flags := getGoLangCILintFlags(configFile, sett.Fix)
	cmd := append([]string{"run"}, flags...)
	cmd = append(cmd, config.Args...)
	cmd = append(cmd, "./...")

	return lintctx.Check(cmd...)
}
