//go:build test && integration

package main

import (
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	configcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/config"
	exrunner "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/exec-runner"
	extarget "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/exec-target"
	listcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/list"
	processcompose "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/process-compose"
	rootcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/root"
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component/query"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/dag"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	execrunnner "github.com/sdsc-ordes/quitsh/pkg/runner/exec"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"
	echorunner "github.com/sdsc-ordes/quitsh/test/runners/echo_test"
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

	// Exec Arguments.
	ExecArgs dag.ExecArgs `yaml:"execArgs"`
}

type Config struct {
	// All command arguments of our `quitsh` instance.
	Commands CommandArgs `yaml:"commands"`

	// Here you can place your own additional global config stuff
	Build settings.BuildSettings `yaml:"build"`

	// A simple test settings which gets env. replaced.
	ValWithEnv string `yaml:"valWithEnv"`
}

func (c *Config) Clone() config.IConfig {
	v, _ := clone.Clone(c).(*Config)

	return v
}

func (c *Config) ExpandEnv() error {
	c.ValWithEnv = os.ExpandEnv(c.ValWithEnv)
	log.Info("Replaced ValWithEnv: '%v'", c.ValWithEnv)

	return nil
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
		// Ignore component-b by not searching in this directory.
		cli.WithCompFindOptions(
			query.WithFindOptions(
				fs.WithWalkDirFilterPatterns(nil, []string{"**/component-b"}, true))),
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
	exrunner.AddCmd(cli, cli.RootCmd(), &args.Commands.DispatchArgs)
	extarget.AddCmd(cli, cli.RootCmd(), &args.Commands.ExecArgs)
	configcmd.AddCmd(cli.RootCmd(), &args)
	listcmd.AddCmd(cli, cli.RootCmd())
	processcompose.AddCmd(cli, cli.RootCmd(), flakeDir)

	// Register the common cmd runner.
	err = execrunnner.Register(
		args.Build.WrapToIBuildSettings(),
		cli.RunnerFactory(), true)

	if err != nil {
		log.PanicE(err, "Could not register runners.")
	}

	// Register some Go runner.
	err = gorunner.Register(&args.Build, cli.RunnerFactory())
	if err != nil {
		log.PanicE(err, "Could not register runners.")
	}

	// Register some other dummy runner.
	err = echorunner.Register(&args.Build, cli.RunnerFactory())
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
