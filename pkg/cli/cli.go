package cli

import (
	rootcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/root"
	"github.com/sdsc-ordes/quitsh/pkg/cli/general"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/component/query"
	"github.com/sdsc-ordes/quitsh/pkg/component/stage"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/runner/factory"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"

	"github.com/spf13/cobra"
)

type ICLI interface {
	// The root directory from where certain operations are done,
	// i.e finding components etc.
	RootDir() string

	// RootCmd returns the root command on the CLI.
	RootCmd() *cobra.Command

	// RootArgs returns the root arguments for the CLI.
	RootArgs() *rootcmd.Args

	// Stages returns the set stage.
	Stages() stage.Stages

	// Config returns the overall customized config for the CLI.
	Config() config.IConfig

	// RunnerFactory gets the runner factory.
	RunnerFactory() factory.IFactory

	// toolchainDispatcher gets the toolchain dispatch function.
	ToolchainDispatcher() toolchain.IDispatcher

	// FindComponents returns components `comps` found by arguments `args` and
	// and all searched components `all` (needed to construct the DAG).
	FindComponents(args *general.ComponentArgs) (
		comps []*component.Component,
		all []*component.Component,
		rootDir string,
		err error,
	)

	// Run will run the CLI.
	Run() error
}

// New creates a new `quitsh` CLI application.
// The root arguments `args` generally point into the `config`.
// The CLI instance needs the full `config`
// because it will marshall/unmarshall it
// from disk by the `rootCmd`.
func New(args *rootcmd.Args, config config.IConfig, opts ...Option) (ICLI, error) {
	app := &cliApp{
		rootArgs: args,
		config:   config,
	}

	for i := range opts {
		if e := opts[i](app); e != nil {
			return nil, e
		}
	}

	if len(app.stages) == 0 {
		app.stages = stage.NewDefaults()
	}

	app.factory = factory.NewFactory(app.Stages())

	app.rootCmd, app.rootCmdPreExec =
		rootcmd.New(&app.settings, app.rootArgs, app.config)

	return app, nil
}

type cliApp struct {
	rootCmd        *cobra.Command
	rootCmdPreExec func() error

	rootDirResolved bool
	rootArgs        *rootcmd.Args

	configFilename string
	config         config.IConfig

	settings rootcmd.Settings

	compFindOpts []query.Option

	stages                  stage.Stages
	targetNameToStageMapper stage.TargetNameToStageMapper

	factory             factory.IFactory
	toolchainDispatcher toolchain.IDispatcher
}
