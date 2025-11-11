package config

import (
	"github.com/creasty/defaults"
	gclone "github.com/huandu/go-clone/generic"
	rootcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/root"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/dag"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"

	cconfig "quitsh-cli/pkg/runner/config"
)

type CommandArgs struct {
	// Arguments needed to make the root command in `quitsh` work.
	Root rootcmd.Args `yaml:"root"`

	// Arguments needed to make the `execute`
	// command in `quitsh` work. This is used when `quitsh` dispatches over a toolchain
	// and needs to call it self (see `exec.AddCmd`).
	DispatchArgs toolchain.DispatchArgs `yaml:"toolchainDispatch"`

	// Exec Arguments.
	ExecArgs dag.ExecArgs `yaml:"execArgs"`
}

type Config struct {
	// All command arguments of our `quitsh` instance.
	Commands CommandArgs `yaml:"commands"`

	// The build settings which get copied and injected into the runners:
	Build cconfig.BuildSettings `yaml:"build"`

	// The lint settings which get copied and injected into the runners:
	// - `custodian::lint-go`
	Lint cconfig.LintSettings `yaml:"lint"`

	// The test settings which get copied and injected into the runners:
	Test cconfig.TestSettings `yaml:"test"`
}

// New returns custodians arguments with default values.
func New() (args Config) {
	err := defaults.Set(&args)
	log.PanicE(err, "could not default initialize config")

	// Fields which are also flags will be initialized
	// by the flags default values.
	return
}

// Clone implements `cli.IConfig` interface.
func (c *Config) Clone() config.IConfig {
	return gclone.Clone(c)
}
