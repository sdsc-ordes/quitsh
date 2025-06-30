package gorunner

import (
	"os"
	"path"

	cm "github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	gox "github.com/sdsc-ordes/quitsh/pkg/exec/go"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const GoTestBinRunnerID = "quitsh::test-go-bin"

type GoTestBinRunner struct {
	runnerConfig *RunnerConfigTestBin
	settings     config.ITestSettings
}

// NewGoTestBinRunner creates a runner which builds an instrumented Go binary
// and tests it with Go tests.
func NewGoTestBinRunner(config any, settings config.ITestSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &GoTestBinRunner{
		runnerConfig: cm.Cast[*RunnerConfigTestBin](config),
		settings:     settings,
	}, nil
}

func (r *GoTestBinRunner) ID() runner.RegisterID {
	return GoTestBinRunnerID
}

func buildBinary(
	log log.ILog,
	comp *component.Component,
	setts config.ITestSettings,
	modInfo gox.GoModInfo,
	runnerConf *RunnerConfigTestBin,
	outputDir string,
) error {
	log.Info("Build instrumented binaries.")

	goctx := gox.NewCtxBuilder().
		Cwd(comp.Root()).
		Env(os.Environ()...).
		Env("GOBIN="+outputDir,
			"GOWORK=off").
		Build()

	flags, tagArgs := GetBuildFlags(
		log,
		comp.Root(),
		setts.BuildType(),
		cm.EnvironmentDev,
		true,
		setts.ShowTestLog(),
		modInfo,
		comp.Version(),
		runnerConf.VersionModule,
		runnerConf.BuildTags,
		true,
	)

	log.Info("Run Go generate.")
	cmd := append([]string{"generate"}, tagArgs...)
	cmd = append(cmd, "./...")
	err := goctx.Check(cmd...)
	if err != nil {
		log.ErrorE(err, "Go generate failed.")

		return err
	}

	log.Info("Run Go install.")
	cmd = append([]string{"install"}, flags...)
	cmd = append(cmd, path.Join(comp.Root(), runnerConf.BuildPkg, "..."))
	err = goctx.Check(cmd...)

	if err != nil {
		log.ErrorE(err, "Go build binaries instrumented failed.")

		return err
	}

	return nil
}

func testBinary(
	log log.ILog,
	comp *component.Component,
	setts config.ITestSettings,
	modInfo gox.GoModInfo,
	runnerConf *RunnerConfigTestBin,
) error {
	envs := []string{
		"GOWORK=off",
		"GOTOOLCHAIN=local",
		"QUITSH_BIN_DIR=" + comp.OutCoverageBinDir(),
		"QUITSH_COVERAGE_DIR=" + comp.OutCoverageDataDir()}
	goctx := gox.NewCtxBuilder().
		Cwd(comp.Root()).
		Env(envs...).
		Build()

	flags, _ := GetBuildFlags(
		log,
		comp.Root(),
		setts.BuildType(),
		cm.EnvironmentDev,
		false,
		setts.ShowTestLog(),
		modInfo,
		comp.Version(),
		runnerConf.VersionModule,
		runnerConf.TestTags,
		true,
	)

	log.Info("Run Go test for testing binary.", "env", envs)
	cmd := append([]string{"test"}, flags...)
	cmd = append(cmd, setts.Args()...)
	cmd = append(cmd, path.Join(comp.Root(), runnerConf.TestPkg, "..."))
	err := goctx.Check(cmd...)

	if err != nil {
		log.ErrorE(err, "Go bin test failed.")

		return err
	}

	return nil
}

func (r *GoTestBinRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	comp := ctx.Component()
	config := comp.Config()
	log.Info("Starting Go bin test for component.", "component", config.Name)

	if len(r.runnerConfig.TestTags) == 0 {
		return errors.New("you must provide at least one tag in `testTags`")
	}

	covDataDir := comp.OutCoverageDataDir()
	fs.AssertDirs(comp.OutCoverageBinDir(), covDataDir)

	modInfo, err := gox.GetModuleInfo(comp.Root())
	if err != nil {
		return err
	}

	err = buildBinary(
		log,
		comp,
		r.settings,
		modInfo,
		r.runnerConfig,
		comp.OutCoverageBinDir(),
	)
	if err != nil {
		log.ErrorE(err, "Building instrumented binary failed.")

		return err
	}

	err = testBinary(
		log,
		comp,
		r.settings,
		modInfo,
		r.runnerConfig,
	)
	if err != nil {
		log.ErrorE(err, "Testing instrumented binary failed.")

		return err
	}

	err = generateCoverageReport(log, comp)
	if err != nil {
		log.ErrorE(err, "Generating coverage report failed.")
	}

	return err
}
