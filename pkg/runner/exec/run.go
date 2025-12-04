package execrunner

import (
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const ExecRunnerID = "quitsh::exec"

type ExecRunner struct {
	config   *RunnerConfig
	settings config.IBuildSettings
}

func NewExecRunner(config any, settings config.IBuildSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &ExecRunner{
		config:   common.Cast[*RunnerConfig](config),
		settings: settings,
	}, nil
}

func (*ExecRunner) ID() runner.RegisterID {
	return ExecRunnerID
}

func (r *ExecRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	comp := ctx.Component()

	fs.AssertDirs(comp.OutBuildBinDir())

	cmdCtx := exec.NewCmdCtxBuilder().
		Cwd(comp.Root()).
		Env(r.config.Env...).
		Env(comp.OutEnvVariables()...).
		Env(
			"QUITSH_BUILD_TYPE="+r.settings.BuildType().String(),
			"QUITSH_ENVIRONMENT_TYPE="+r.settings.EnvironmentType().String()).
		Build()

	if r.config.Script != "" {
		cmdCtx.WithStdin(strings.NewReader(r.config.Script))
	}

	log.Info(
		"Executing exec runner.",
		"component", comp.Name(),
		"name", r.config.Name)
	log.Debug("Command config", "config", r.config)

	return cmdCtx.Check(r.config.Cmd...)
}
