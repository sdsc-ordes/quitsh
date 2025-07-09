// NOTE: This is `quitsh`s own application of itself!

package main

import (
	"os"

	cliconfig "quitsh-cli/config"
	cliGoRunner "quitsh-cli/pkg/runner/go"

	"github.com/sdsc-ordes/quitsh/pkg/cli"
	configcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/config"
	execrunner "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/exec-runner"
	exectarget "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/exec-target"
	listcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/list"
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component/query"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	gorunner "github.com/sdsc-ordes/quitsh/pkg/runner/go"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"
)

func main() {
	err := log.Setup("info") // Level will be set at startup.
	if err != nil {
		log.PanicE(err, "Could not setup logger.")
	}

	conf := cliconfig.New()

	cli, err := cli.New(
		&conf.Commands.Root,
		&conf,
		cli.WithName("cli"),
		cli.WithDescription("This is the 🐔-🥚 CLI tool for 'quitsh', yes its built with 'quitsh'."),
		cli.WithCompFindOptions(
			query.WithFindOptions(
				fs.WithWalkDirFilterPatterns(nil,
					[]string{"**/test/repo/**"}, true))),
		cli.WithStages("lint", "build", "test"),
		cli.WithTargetToStageMapperDefault(),
		cli.WithToolchainDispatcherNix(
			"tools/nix",
			func(c config.IConfig) *toolchain.DispatchArgs {
				cc := common.Cast[*cliconfig.Config](c)

				return &cc.Commands.DispatchArgs
			},
		),
	)
	if err != nil {
		log.PanicE(err, "Could not initialize CLI app.")
	}

	// Setup quitsh provided helper commands.
	listcmd.AddCmd(cli, cli.RootCmd())
	configcmd.AddCmd(cli.RootCmd(), &conf)
	exectarget.AddCmd(cli, cli.RootCmd(), &conf.Commands.ExecArgs)
	execrunner.AddCmd(cli, cli.RootCmd(), &conf.Commands.DispatchArgs)

	registerRunners(cli, &conf)

	// Run the app.
	err = cli.Run()
	if err != nil {
		log.ErrorE(err, "Error occurred.")
		os.Exit(1)
	}
}

func registerRunners(cl cli.ICLI, args *cliconfig.Config) {
	err := gorunner.RegisterBuild(args.Build.WrapToIBuildSettings(), cl.RunnerFactory(), true)
	if err != nil {
		log.PanicE(err, "Could not register runner.")
	}

	err = gorunner.RegisterTest(args.Test.WrapToITestSettings(), cl.RunnerFactory(), true)
	if err != nil {
		log.PanicE(err, "Could not register runner.")
	}

	err = cliGoRunner.Register(&args.Lint, cl.RunnerFactory())
	if err != nil {
		log.PanicE(err, "Could not register runner.")
	}
}
