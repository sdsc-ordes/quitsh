package config

import (
	"github.com/creasty/defaults"
)

type (
	NixSettings struct {
		FlakeDirRel string `yaml:"flakeDirRel" default:"tools/nix"`

		// The host name of the Nix cache for upload and download.
		Cache NixCache `yaml:"cache"`
	}

	NixCache struct {
		SSH NixSSH `yaml:"ssh"`
	}

	NixSSH struct {
		Enable bool `yaml:"enable"`

		HostName      string `yaml:"hostName" validate:"required_if=Enable true"`
		HostPublicKey string `yaml:"hostPublicKey" validate:"required_if=Enable true"`

		// The write user credentials.
		// This will use the `ssh-add` over the agent.
		Write NixSSHCredentials `yaml:"write"`

		// The read user credentials.
		// This will not use the `ssh-add` over the agent.
		// but just rely on `~/.ssh/config`.
		Read NixSSHCredentials `yaml:"read"`
	}

	NixSSHCredentials struct {
		User string `yaml:"user" default:"nix-ssh"`

		// If the keys are added to the agent or only the `~/.ssh/config` is
		// necessary.
		UseAgent       bool   `yaml:"agent"`
		PrivateKeyEnv  string `yaml:"privateKeyEnv"`
		PrivateKeyPath string `yaml:"privateKeyPath" validate:"required_if=Enable true privateKeyEnv ''"`
	}
)

// Interface implementation guard.
var _ defaults.Setter = (*NixSettings)(nil)

// SetDefaults implements [defaults.Setter].
func (c *NixSettings) SetDefaults() {
	c.Cache.SSH.Write.User = "nix-ssh-write"
	c.Cache.SSH.Write.UseAgent = true
}

func (s *NixSSH) URL(write bool) string {
	if write {
		return "ssh://" + s.Write.User + "@" + s.HostName
	}

	return "ssh://" + s.Read.User + "@" + s.HostName
}
