package trivyrunner

import (
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const TrivyLintRunnerID = "quitsh::lint-trivy"

type (
	TrivyLintRunner struct {
		runnerConfig *RunnerConfigLint
		settings     *config.LintSettings
	}
)

func NewTrivyLintRunner(config any, settings *config.LintSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &TrivyLintRunner{
		runnerConfig: common.Cast[*RunnerConfigLint](config),
		settings:     settings,
	}, nil
}

// Interface implementation guard.
var _ runner.IRunner = (*TrivyLintRunner)(nil)

// ID implements [runner.IRunner].
func (r *TrivyLintRunner) ID() runner.RegisterID {
	return TrivyLintRunnerID
}

// Run implements [runner.IRunner].
func (r *TrivyLintRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	comp := ctx.Component()

	configFile := path.Join(ctx.Root(), r.runnerConfig.ConfigFile)
	if !fs.Exists(configFile) {
		return errors.New("Config file '%v' is not existing.", configFile)
	}

	trivyCtx := exec.NewCmdCtxBuilder().
		Cwd(comp.Root()).BaseCmd("trivy").Build()

	securityIssues := false

	err := trivyCtx.CheckWithEC(
		func(cmdError *exec.CmdError) error {
			securityIssues = cmdError != nil && cmdError.ExitCode() == 114 //nolint:mnd

			return cmdError
		},
		"fs",
		"--exit-code", "114",
		"--config", configFile,
		".",
	)

	if securityIssues {
		log.Error("Security issues are found, please check trivy output above.")

		return errors.New("security issues detected")
	} else if err != nil {
		return err
	}

	return err
}
