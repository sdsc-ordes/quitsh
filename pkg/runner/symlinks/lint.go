package symlinkrunner

import (
	stdfs "io/fs"
	"os"
	"path/filepath"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/debug"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/runner"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"
)

const SymlinkLintRunnerID = "quitsh::lint-symlinks"

type (
	SymlinkLintRunner struct {
		settings     *config.LintSettings
		runnerConfig *SymlinkConfig
	}
)

func NewSymlinkLintRunner(config any, settings *config.LintSettings) (runner.IRunner, error) {
	debug.Assert(config != nil, "config is nil")

	return &SymlinkLintRunner{
		runnerConfig: common.Cast[*SymlinkConfig](config),
		settings:     settings,
	}, nil
}

// Interface implementation guard.
var _ runner.IRunner = (*SymlinkLintRunner)(nil)

// ID implements [runner.IRunner].
func (r *SymlinkLintRunner) ID() runner.RegisterID {
	return SymlinkLintRunnerID
}

// Run implements [runner.IRunner].
func (r *SymlinkLintRunner) Run(ctx runner.IContext) error {
	log := ctx.Log()
	gitx := ctx.Git()

	files, traversedFiles, _ := gitx.FilterFiles(
		fs.WithPathFilter(func(_path string, info os.DirEntry) bool {
			return info.Type()&stdfs.ModeSymlink != 0
		}, true),
	)

	log.Infof("Found '%v' symlinks in '%v' files.", len(files), traversedFiles)

	var err error
	for i := range files {
		f, e := filepath.EvalSymlinks(files[i])
		f = filepath.ToSlash(f)

		if e != nil {
			link, e2 := os.Readlink(files[i])
			err = errors.Combine(err, e, e2,
				errors.New("⛑️  symlink file '%v' is broken (points to: '%v')", files[i], link))
		} else if !fs.Exists(f) {
			err = errors.Combine(
				err,
				errors.New(
					"⛑️  symlink file '%v' points to non-existing file ('%v').",
					files[i],
					f,
				),
			)
		}

		log.Debugf("✔️ '%v' symlink is ok.", files[i])
	}

	return err
}
