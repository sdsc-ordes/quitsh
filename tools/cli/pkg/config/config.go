package config

import (
	cnConfig "quitsh-cli/pkg/runner/config"

	"github.com/creasty/defaults"
	rootcmd "github.com/sdsc-ordes/quitsh/pkg/cli/cmd/root"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/toolchain"

	"github.com/huandu/go-clone"
)

type CommandArgs struct {
	// Arguments needed to make the root command in `quitsh` work.
	Root rootcmd.Args `yaml:"general"`

	// Arguments needed to make the `execute`
	// command in `quitsh` work. This is used when `quitsh` dispatches over a toolchain
	// and needs to call it self (see `exec.AddCmd`).
	DispatchArgs toolchain.DispatchArgs `yaml:"toolchainDispatch"`
}

type Config struct {
	// All command arguments of our `quitsh` instance.
	Commands CommandArgs `yaml:"commands"`

	// The build settings which get copied and injected into the runners:
	// - `custodian::build-go`
	Build cnConfig.BuildSettings `yaml:"build"`

	// The lint settings which get copied and injected into the runners:
	// - `custodian::lint-go`
	Lint cnConfig.LintSettings `yaml:"lint"`

	// The test settings which get copied and injected into the runners:
	// - `custodian::test-go`
	Test cnConfig.TestSettings `yaml:"test"`
}

// New returns custodians arguments with default values.
func New() (args Config) {
	// Fields which are also flags will be initialized
	// by the flags default values.

	err := defaults.Set(&args)
	log.PanicE(err, "could not default initialize config")

	return
}

// Implement `cli.IConfig` interface.
func (c *Config) Clone() config.IConfig {
	v, _ := clone.Clone(c).(*Config)

	return v
}
