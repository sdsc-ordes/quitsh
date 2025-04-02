//go:build test && integration

package main

import (
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	execrunner "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/exec-runner"
	exectarget "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/exec-target"
	listcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/list"
	processcompose "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/process-compose"
	rootcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/root"
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"
	gorunner "github.com/sdsc-ordes/quitsh/test/runners/go_test"
	settings "github.com/sdsc-ordes/quitsh/test/runners/settings_test"

	"github.com/huandu/go-clone"
)

type CommandArgs struct {
	// Arguments needed to make the root command in `quitsh` work.
	Root rootcmd.Args `yaml:"general"`

	// Arguments needed to make the `execute`
	// command in `quitsh` work. This is used when `quitsh` dispatches over a toolchain
	// and needs to call it self (see `exec.AddCmd`).
	DispatchArgs toolchain.DispatchArgs `yaml:"toolchainDispatch"`
}

type Config struct {
	// All command arguments of our `quitsh` instance.
	Commands CommandArgs `yaml:"commands"`

	// Here you can place your own additional global config stuff
	Build settings.BuildSettings `yaml:"build"`
}

// Implement `cli.IConfig` interface.
func (c *Config) Clone() config.IConfig {
	v, _ := clone.Clone(c).(*Config)

	return v
}

func main() {
	err := log.Setup("info") // Level will be set at startup.
	if err != nil {
		log.PanicE(err, "Could not setup logger.")
	}

	var args Config

	flakeDir := "."
	cli, err := cli.New(
		&args.Commands.Root,
		&args,
		cli.WithName("custodian-test"),
		cli.WithStages("lint", "build", "test", "monkey-stage", "deploy"),
		cli.WithTargetToStageMapperDefault(),
		cli.WithToolchainDispatcherNix(
			flakeDir,
			func(c config.IConfig) *toolchain.DispatchArgs {
				cc := common.Cast[*Config](c)

				return &cc.Commands.DispatchArgs
			},
		),
	)
	if err != nil {
		log.PanicE(err, "Could not setup cli.")
	}

	// Setup quitsh provided helper commands.
	execrunner.AddCmd(cli, cli.RootCmd(), &args.Commands.DispatchArgs)
	exectarget.AddCmd(cli, cli.RootCmd())
	listcmd.AddCmd(cli, cli.RootCmd())
	processcompose.AddCmd(cli, cli.RootCmd(), flakeDir)

	// Register some Go runner.
	err = gorunner.Register(&args.Build, cli.RunnerFactory())
	if err != nil {
		log.PanicE(err, "Could not register runners.")
	}

	// Run the app.
	err = cli.Run()
	if err != nil {
		log.ErrorE(err, "Error occurred.")
		os.Exit(1)
	}
}
