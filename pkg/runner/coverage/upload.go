package coverage

import (
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/component/step"
	"github.com/sdsc-ordes/quitsh/pkg/coverage"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const CoverageUploadRunnerID = "quitsh::coverage-upload"

type (
	CoverageUpload struct {
		runnerConfig *RunnerConfigCoverageUpload
		settings     config.ITestSettings
	}

	RunnerConfigCoverageUpload struct {
	}
)

func UnmarshalCodecovConfig(_raw step.AuxConfigRaw) (step.AuxConfig, error) {
	return &RunnerConfigCoverageUpload{}, nil
}

func NewCodecovRunner(config any, settings config.ITestSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &CoverageUpload{
		runnerConfig: common.Cast[*RunnerConfigCoverageUpload](config),
		settings:     settings,
	}, nil
}

// Interface implementation guard.
var _ runner.IRunner = (*CoverageUpload)(nil)

// ID implements [runner.IRunner].
func (r *CoverageUpload) ID() runner.RegisterID {
	return CoverageUploadRunnerID
}

// Run implements [runner.IRunner].
func (r *CoverageUpload) Run(ctx runner.IContext) error {
	log := ctx.Log()

	if !ci.IsRunning() || os.Getenv("NIX_BUILD_TOP") != "" {
		log.Info("CI is not running or inside Nix build, coverage upload skipped.")

		return nil
	}

	comp := ctx.Component()

	codecovCtx := exec.NewCmdCtxBuilder().
		Cwd(ctx.Root()).
		BaseCmd("codecov").
		CredentialFilter(nil).
		Build()

	gitx := ctx.Git()
	commitSHA, err := gitx.CurrentRev()
	if err != nil {
		return err
	}

	info := coverage.NewCoverageInfo()
	info.FailIfNoFiles = true
	info.CommitSHA = commitSHA
	info.Flag = comp.Name()
	info.RepoRoot = ctx.Root()

	err = info.AddComponentDefaultFiles(comp)
	if err != nil {
		return err
	}

	return coverage.UploadCoverageCodecov(log, codecovCtx, &info)
}
