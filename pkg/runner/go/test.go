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
	config   *RunnerTestConfig
	settings config.ITestSettings
}

func NewGoTestRunner(config any, settings config.ITestSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &GoTestRunner{
		config:   cm.Cast[*RunnerTestConfig](config),
		settings: settings,
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
		log.Warn(
			"If you use the `submodules` runner config, make sure to have a 'go.work' file " +
				"also to list all submodules to make the HTML conversion resolve paths.",
		)
	}

	return err
}

func (r *GoTestRunner) Run(ctx runner.IContext) error {
	comp := ctx.Component()
	log := ctx.Log()

	config := comp.Config()
	log.Info("Starting Go test for component.", "component", config.Name)

	goctx := gox.NewCtxBuilder().Cwd(comp.Root()).
		Env("GOWORK="+r.config.GoWork,
			"GOTOOLCHAIN="+r.config.GoToolchain).
		Build()

	covDataDir := comp.OutCoverageDataDir()
	fs.AssertDirs(comp.OutBuildBinDir(), covDataDir)

	//FIXME: Somehow we have a cache write/exec race problem with the
	//       `go tool covdata` binary.
	//       Ref: https://github.com/golang/go/issues/78777
	//       Run dummy command to make the tool cached...
	log.Info("Cache 'go tool covdata'. (workaround).")
	_, err := goctx.Get("tool", "covdata", "pkglist", "-i", ".")
	if err != nil {
		return err
	}

	if len(r.config.Submodules) == 0 {
		r.config.Submodules = append(r.config.Submodules, ".")
	}

	for _, submodule := range r.config.Submodules {
		moduleRoot := path.Join(comp.Root(), submodule)

		modInfo, e := gox.GetModuleInfo(moduleRoot)
		if e != nil {
			return e
		}

		flags, _, genArgs := GetBuildFlags(
			log,
			moduleRoot,
			r.settings.BuildType(),
			cm.EnvironmentDev,
			true,
			r.settings.ShowTestLog(),
			modInfo,
			comp.Version(),
			"",
			r.config.BuildTags,
			true,
		)

		log.Infof("Run Go generate for '%v'.", submodule)
		cmd := append([]string{goGenerate}, genArgs...)
		cmd = append(cmd, "./...")
		e = goctx.Check(cmd...)
		if e != nil {
			log.ErrorE(e, "Go generate failed.")

			return e
		}

		log.Info("Run Go test.")
		cmd = append([]string{"test"}, flags...)
		cmd = append(cmd, r.config.Args...)
		cmd = append(cmd, r.settings.Args()...)
		cmd = append(cmd, path.Join(moduleRoot, "..."))
		cmd = append(cmd, "-args", "-test.gocoverdir="+covDataDir)
		cmd = append(cmd, r.config.TestArgs...)
		cmd = append(cmd, r.settings.TestArgs()...)
		e = goctx.Check(cmd...)

		if e != nil {
			log.ErrorE(e, "Go test failed.")

			return e
		}
	}

	err = generateCoverageReport(log, comp)
	if err != nil {
		log.ErrorE(err, "Go coverage conversion failed.")
	}

	return err
}
