package component

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"

	"github.com/hashicorp/go-version"
)

type Version version.Version

// Implement the pflag.Value interface.
func (cv *Version) String() string {
	return (*version.Version)(cv).String()
}

// Implement the pflag.Value interface.
func (cv *Version) Set(s string) error {
	p := (*version.Version)(cv)

	err := p.UnmarshalText([]byte(s))
	if err != nil {
		return errors.AddContext(err, "version '%v' is not a sem. version", s)
	}

	return nil
}

// Implement the pflag.Value interface.
func (cv *Version) Type() string {
	return "ComponentVersion"
}

func (cv *Version) UnmarshalText(bytes []byte) error {
	p := (*version.Version)(cv)

	return p.UnmarshalText(bytes)
}
