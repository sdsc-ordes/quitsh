package gorunner

import (
	"path"

	cm "github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	gox "github.com/sdsc-ordes/quitsh/pkg/exec/go"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const GoTestRunnerID = "quitsh::test-go"

type GoTestRunner struct {
	runnerConfig *RunnerConfigBuild
	settings     config.ITestSettings
}

func NewGoTestRunner(config any, settings config.ITestSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &GoTestRunner{
		runnerConfig: cm.Cast[*RunnerConfigBuild](config),
		settings:     settings,
	}, nil
}

func (r *GoTestRunner) ID() runner.RegisterID {
	return GoTestRunnerID
}

func generateCoverageReport(log log.ILog, comp *component.Component) error {
	covDataDir := comp.OutCoverageDataDir()
	covHTML := comp.OutCoverageDataDir("coverage.html")
	covFile := comp.OutCoverageDataDir("coverage.txt")
	log.Info("Generating coverage file.", "path", "file://"+covHTML)

	goctx := gox.NewCtxBuilder().Cwd(comp.Root()).Build()

	err := goctx.Chain().
		Check("tool", "covdata", "textfmt", "-i", covDataDir, "-o", covFile).
		Check("tool", "cover", "-html="+covFile, "-o", covHTML).
		Error()

		// TODO: Add gocover-cobertura conversion to upload to Gitlab
		// See Issue: https://gitlab.com/data-custodian/custodian/-/issues/196
	if err != nil {
		log.ErrorE(err, "Go coverage conversion failed.")
	}

	return err
}

func (r *GoTestRunner) Run(ctx runner.IContext) error {
	comp := ctx.Component()
	log := ctx.Log()

	config := comp.Config()
	log.Info("Starting Go test for component.", "component", config.Name)

	goctx := gox.NewCtxBuilder().Cwd(comp.Root()).
		Env("GOWORK=off",
			"GOTOOLCHAIN=local").
		Build()

	covDataDir := comp.OutCoverageDataDir()
	fs.AssertDirs(comp.OutBuildBinDir(), covDataDir)

	modInfo, err := gox.GetModuleInfo(comp.Root())
	if err != nil {
		return err
	}

	log.Info("Run Go generate.")
	err = goctx.Check("generate", "./...")
	if err != nil {
		log.ErrorE(err, "Go generate failed.")

		return err
	}

	flags := GetBuildFlags(
		comp.Root(),
		r.settings.BuildType(),
		cm.EnvironmentDev,
		true,
		r.settings.ShowTestLog(),
		modInfo,
		comp.Version(),
		r.runnerConfig.VersionModule,
		r.runnerConfig.BuildTags,
		true,
	)

	// TODO: Run `go test` over `grc --config root_dir/tools/config/grc/...` to colorize.
	//       Issue: https://gitlab.com/data-custodian/custodian/-/issues/194
	log.Info("Run Go test.")
	cmd := append([]string{"test"}, flags...)
	cmd = append(cmd, r.settings.Args()...)
	cmd = append(cmd, path.Join(comp.Root(), "..."))
	cmd = append(cmd, "-args", "-test.gocoverdir="+covDataDir)
	err = goctx.Check(cmd...)

	if err != nil {
		log.ErrorE(err, "Go test failed.")

		return err
	}

	err = generateCoverageReport(log, comp)
	if err != nil {
		log.ErrorE(err, "Go coverage conversion failed.")
	}

	return err
}
