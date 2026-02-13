//go:build test

package gorunner

import (
	"github.com/sdsc-ordes/quitsh/pkg/runner"
)

const GoLintRunnerID = "integration-test::lint-go"

type GoLintRunner struct {
}

// NewGoLintRunner constructs a new GoLintRunner with its own config.
func NewGoLintRunner(config any) (runner.IRunner, error) {
	return &GoLintRunner{}, nil
}

func (*GoLintRunner) ID() runner.RegisterID {
	return GoLintRunnerID
}

func (r *GoLintRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	comp := ctx.Component()

	log.Info("Hello from integration test Go lint runner.", "component", comp.Name())
	return nil
}
