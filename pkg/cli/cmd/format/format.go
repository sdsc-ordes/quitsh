package format

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/ci"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec/git"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	fs "github.com/sdsc-ordes/quitsh/pkg/filesystem"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/runner/config"

	"github.com/spf13/cobra"
)

const longDesc = `
Format the repository with 'nix run treefmt' ->.
`

type FormatArgs struct {
	FailOnChange bool
	CI           bool
}

func AddCmd(root *cobra.Command, nixSetts *config.NixSettings) {
	var fmtArgs FormatArgs

	formatCmd := &cobra.Command{
		Use:   "format [paths...]",
		Short: "Format the files in paths (default '.')",
		Long:  longDesc,
		RunE: func(_cmd *cobra.Command, args []string) error {
			_, rootDir, err := git.NewCtxAtRoot(".")
			if err != nil {
				return err
			}

			flakeDir := path.Join(rootDir, nixSetts.FlakeDirRel)

			return FormatPaths(&fmtArgs, args, rootDir, flakeDir)
		},
	}

	formatCmd.Flags().
		BoolVarP(&fmtArgs.FailOnChange,
			"fail-on-change", "c", false, "Fail if any changes were made.")

	formatCmd.Flags().
		BoolVar(&fmtArgs.CI,
			"ci", ci.IsRunning(), "Run format in CI. (no cache, fail on change)")

	root.AddCommand(formatCmd)
}

// FormatPaths formats all paths (or `rootDir` by default)
// over `treefmt`.
func FormatPaths(
	c *FormatArgs,
	paths []string,
	rootDir string,
	flakeDir string) error {
	if len(paths) == 0 {
		paths = append(paths, rootDir)
	}

	// Make all paths relative to the root dir.
	files := []string{}
	for i := range paths {
		if paths[i] == "" {
			continue
		}

		if strings.HasPrefix(paths[i], "-") {
			continue
		}

		absP := fs.MakeAbsolute(paths[i])
		if !fs.Exists(absP) {
			return errors.New("path '%s' does not exist'", absP)
		}

		relPath, e := filepath.Rel(rootDir, absP)
		if e != nil {
			return e
		}

		files = append(files, relPath)
	}

	log.Infof("Formatting '%v' paths.", len(files))
	log.Debug("Files", "files", files)

	nixctx := nix.NewRunCtx(rootDir)

	cmd := []string{flakeDir + "#treefmt"}
	cmd = append(cmd, "--")
	cmd = append(cmd, files...)
	if c.FailOnChange {
		cmd = append(cmd, "--fail-on-change")
	}

	if c.CI {
		cmd = append(cmd, "--ci")
	}

	return nixctx.Check(cmd...)
}
