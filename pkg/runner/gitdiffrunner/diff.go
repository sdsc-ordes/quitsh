package gitdiffrunner

import (
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/cli/cmd/format"
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const GitDiffRunnerID = "quitsh::git-diff"

type GitDiffRunner struct {
	config   *RunnerDiffConfig
	nixSetts *config.NixSettings
}

func NewGitDiffRunner(config any, nixSetts *config.NixSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &GitDiffRunner{
		config:   common.Cast[*RunnerDiffConfig](config),
		nixSetts: nixSetts,
	}, nil
}

// Interface implementation guard.
var _ runner.IRunner = (*GitDiffRunner)(nil)

// ID implements [runner.IRunner].
func (*GitDiffRunner) ID() runner.RegisterID {
	return GitDiffRunnerID
}

// Run implements [runner.IRunner].
func (r *GitDiffRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	comp := ctx.Component()

	if !ci.IsRunning() {
		log.Info("Skipping Git diff runner, since not running in CI.")

		return nil
	}

	config := comp.Config()
	log.Info("Checking that no diffs are visible.", "component", config.Name)

	gitx := ctx.Git()
	changes, err := getChanges(
		log, gitx, r.config,
		ctx.Root(), comp.Root())
	if err != nil {
		return err
	}

	err = formatFiles(gitx, changes,
		ctx.Root(), r.nixSetts.FlakeDirRel)
	if err != nil {
		return errors.AddContext(err, "could not format all changed files")
	}

	changes, err = getChanges(
		log, gitx, r.config,
		ctx.Root(), comp.Root())
	if err != nil {
		return err
	}

	if len(changes) != 0 {
		log.Error("You have changed or staged files which are not allowed!.")
		log.Error("Changed files.", "files", changes)

		return errors.New("changed files detected which is not allowed")
	}

	return nil
}

func formatFiles(gitx git.Context, changes []string, rootDir, flakeDirRel string) error {
	err := format.FormatPaths(
		&format.FormatArgs{FailOnChange: false, CI: false},
		changes, rootDir,
		path.Join(rootDir, flakeDirRel))
	if err != nil {
		return err
	}

	for i := range changes {
		err = gitx.Check("add", changes[i])
		if err != nil {
			return errors.AddContext(err, "could not stage formatted file '%v'",
				changes[i])
		}
	}

	return nil
}

func getChanges(
	log log.ILog,
	gitx git.Context,
	config *RunnerDiffConfig,
	rootDir, compDir string,
) (changes []string, err error) {
	for _, p := range config.RootPaths {
		f, e := assertNoDiff(log, gitx, path.Join(rootDir, p))
		err = errors.Combine(err, e)
		changes = append(changes, f...)
	}

	for _, p := range config.Paths {
		f, e := assertNoDiff(log, gitx, path.Join(compDir, p))
		err = errors.Combine(err, e)
		changes = append(changes, f...)
	}

	return changes, err
}

func assertNoDiff(log log.ILog, gitx git.Context, path string) ([]string, error) {
	if !fs.Exists(path) {
		return nil, errors.New("path '%s' for diffing does not exist", path)
	}

	log.Infof("Checking no diffs in '%s'.", path)

	staged, err := gitx.GetSplit("diff", "--cached", "--name-only", path)
	if err != nil {
		return nil, errors.AddContext(err, "could not get Git staged files")
	}

	changed, err := gitx.GetSplit("diff", "--name-only", path)
	if err != nil {
		return nil, errors.AddContext(err, "could not get Git changed files")
	}

	return append(changed, staged...), nil
}
