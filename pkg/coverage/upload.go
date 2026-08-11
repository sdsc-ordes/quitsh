package coverage

import (
	"os"
	"path"
	"slices"

	"deedles.dev/xiter"
	"github.com/creasty/defaults"
	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/component"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/secret"
)

// CoverageToolInfo is information for the coverage tool, (currently this is `codecov`).
type CoverageToolInfo struct {
	RepoRoot      string
	CommitSHA     string
	FailIfNoFiles bool
	TokenEnv      string `default:"CODECOV_TOKEN"`

	Files []string
	Flag  string `default:"local"`

	ConfigFile string `default:"tools/configs/codecov/codecov.yaml"`

	DryRun bool
}

// NewCoverageInfo returns a default initialized coverage info.
func NewCoverageInfo() (info CoverageToolInfo) {
	err := defaults.Set(&info)
	log.PanicE(err, "should not happen")

	return
}

// AddComponentDefaultFiles adds the components default coverage
// files currently supported.
func (info *CoverageToolInfo) AddComponentDefaultFiles(comp *component.Component) error {
	coverageFiles := []string{
		path.Join(comp.OutCoverageDataDir("coverage.txt")),
		path.Join(comp.OutCoverageDataDir("coverage-codecov.json")),
	}

	files := slices.Collect(
		xiter.Filter(slices.Values(coverageFiles), fs.Exists),
	)

	if len(files) == 0 {
		log.Warnf("Could not find any coverage files to upload for component '%v'!",
			comp.Name())

		if info.FailIfNoFiles || ci.IsRunning() {
			return errors.New(
				"no coverage files found to upload in '%v', "+
					"candidates are: '%v'",
				comp.OutCoverageDataDir(),
				coverageFiles)
		}
	}

	info.Files = append(info.Files, files...)

	return nil
}

// InitCoverageCodecov will create a report and commit object:
// Its not more needed: See https://github.com/codecov/codecov-cli/issues/696
func InitCoverageCodecov(
	log log.ILog,
	codecovCtx *exec.CmdContext,
	info *CoverageToolInfo,
) error {
	log.Info("Init coverage report.")
	cred, err := secret.NewCredentialsTokenOnly(info.TokenEnv)
	if err != nil {
		return errors.AddContext(err,
			"token not defined, maybe you want to use `--dry-run` instead?")
	}

	getArgs := func(name string) []string {
		cmd := getDefaultArgs(info)
		cmd = append(cmd, name)
		cmd = append(cmd, getRepoArgs()...)

		return append(cmd,
			"--fail-on-error",
			"--token", cred.Token(),
		)
	}

	err = codecovCtx.Check(getArgs("create-commit")...)
	if err != nil {
		return errors.AddContext(err, "could not create commit with `codecov`")
	}

	err = codecovCtx.Check(getArgs("create-report")...)
	if err != nil {
		return errors.AddContext(err, "could not create report with `codecov`")
	}

	return nil
}

// UploadCoverageCodecov uploads coverage with the `codecov` tool.
func UploadCoverageCodecov(
	log log.ILog,
	codecovCtx *exec.CmdContext,
	info *CoverageToolInfo,
) error {
	log.Info("Upload coverage reports.")
	if len(info.Files) == 0 {
		log.Warn("Could not find any coverage files to upload!")
		if info.FailIfNoFiles || ci.IsRunning() {
			return errors.New("no coverage files configured to upload")
		}

		return nil
	}

	cmd := getDefaultArgs(info)

	cred, err := secret.NewCredentialsTokenOnly(info.TokenEnv)
	if err != nil {
		return errors.AddContext(err,
			"token not defined, maybe you want to use `--dry-run` instead?")
	}

	// This command is the current one:
	// See: https://github.com/codecov/codecov-cli/issues/696
	cmd = append(cmd, "upload-coverage")
	cmd = append(cmd, getRepoArgs()...)

	if info.DryRun {
		cmd = append(cmd, "--dry-run")
	}
	cmd = append(cmd,
		"--network-root-folder", info.RepoRoot,
		"--flag", info.Flag,
		"--disable-search",
		"--fail-on-error",
		"--token", cred.Token(),
	)

	log.Info("Codecov info.", "info", info)

	for i := range info.Files {
		cmd = append(cmd, "-f", info.Files[i])
	}

	return codecovCtx.Check(cmd...)
}

func getDefaultArgs(info *CoverageToolInfo) (cmd []string) {
	cmd = append(cmd, "--verbose")

	if ci.IsRunning() {
		switch {
		case os.Getenv("GITLAB_CI") == "true":
			cmd = append(cmd, "--auto-load-params-from", "gitlabci")
		case os.Getenv("GITHUB_ACTIONS") == "true":
			cmd = append(cmd, "--auto-load-params-from", "githubactions")
		default:
			log.Warn("Using default behavior of 'codecov' for loading CI parameters.")
		}
	} else {
		log.Warn("Using default behavior of 'codecov' for loading CI parameters.")
	}

	info.ConfigFile = fs.MakeAbsoluteTo(info.RepoRoot, info.ConfigFile)
	if fs.Exists(info.ConfigFile) {
		cmd = append(cmd, "--codecov-yml-path", info.ConfigFile)
	} else {
		log.Infof("Coverage config file '%s' does not exist. Not using it.", info.ConfigFile)
	}

	return cmd
}

func getRepoArgs() []string {
	// No args are needed since `--auto-load-params-from`
	return nil
}
