package nixtoolchain

import (
	"os"
	"path"

	"github.com/sdsc-ordes/quitsh/pkg/build"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/exec/nix"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"

	"github.com/goccy/go-yaml"
)

type ArgsSelector func(config.IConfig) *toolchain.DispatchArgs

type NixDispatcher struct {
	flakeDirRel string

	command []string

	argsSelector ArgsSelector
}

func storeConfig(config config.IConfig) (file string, cleanup func(), err error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return
	}

	cleanup = func() {
		if !build.DebugEnabled {
			e := os.RemoveAll(dir)
			if e != nil {
				log.WarnE(e, "Could not remove temp dir.")
			}
		}
	}

	file = path.Join(dir, "config.yaml")
	log.Debug("Store config.", "path", file)

	f, err := os.Create(file)
	if err != nil {
		return
	}
	defer f.Close()

	e := yaml.NewEncoder(f)
	defer e.Close()

	err = e.Encode(config)
	if err != nil {
		return
	}

	return file, cleanup, nil
}

func (d *NixDispatcher) Run(
	rootDir string,
	dArgs *toolchain.DispatchArgs,
	config config.IConfig,
) error {
	configCopy := config.Clone()

	// Sett all values.
	args := d.argsSelector(configCopy)
	args.ComponentDir = dArgs.ComponentDir
	args.TargetID = dArgs.TargetID
	args.StepIndex = dArgs.StepIndex
	args.RunnerIndex = dArgs.RunnerIndex
	args.RunnerID = dArgs.RunnerID
	args.Toolchain = dArgs.Toolchain

	file, cleanup, err := storeConfig(configCopy)
	if err != nil {
		return err
	}
	defer cleanup()

	flakePath := path.Join(rootDir, d.flakeDirRel)
	nixToolchainRef := nix.ToolchainInstallable(flakePath, dArgs.Toolchain)
	log.Info("Dispatching to toolchain.", "toolchain", nixToolchainRef)

	// Call the tool again, but over Nix.
	nixctx := NewCtxBuilder(rootDir, flakePath, dArgs.Toolchain).Build()
	nixCmd := append([]string{os.Args[0]}, d.command...)
	nixCmd = append(nixCmd, "--config", file)

	return nixctx.Check(nixCmd...)
}

// NewDispatcher creates a function which dispatches over the tool again, but
// running it with
//
// ```shell
//
//	nix develop "<repoDir>/<flakeDirRel>#<toolchain> \
//	  --command <args-0> <command>...
//	  --config <tempfile.yaml>
//
// ```
// where `<args-0>` is the executable of the quitsh instance,
// and `<command>` some custom arguments,
// and the whole config file `config` is serialized to
// passed as `--config <tempFile.yaml>`.
func NewDispatcher(
	flakeDir string,
	command []string,
	argsSelector ArgsSelector,
) NixDispatcher {
	return NixDispatcher{flakeDir, command, argsSelector}
}
