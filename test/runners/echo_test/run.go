//go:build test

package echorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	setts "github.com/sdsc-ordes/quitsh/test/runners/settings_test"
)

const EchoRunnerID = "integration-test::echo"

type EchoRunner struct {
	runnerConfig *EchoConfig
	settings     *setts.BuildSettings
}

// NewEchoRunner constructs a new GoBuildRunner with its own config.
func NewEchoRunner(config any, settings *setts.BuildSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &EchoRunner{
		runnerConfig: common.Cast[*EchoConfig](config),
		settings:     settings,
	}, nil
}

func (*EchoRunner) ID() runner.RegisterID {
	return EchoRunnerID
}

func (r *EchoRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()

	log.Info("Hello from echo runner", "text", r.runnerConfig.Text)

	return nil
}
