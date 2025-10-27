package component

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"

	"github.com/hashicorp/go-version"
)

type Version struct {
	version.Version
}

// String implements the [pflag.Value] interface.
func (v *Version) String() string {
	return v.Version.String()
}

// Set implements the [pflag.Value] interface.
func (v *Version) Set(s string) error {
	err := v.Version.UnmarshalText([]byte(s))
	if err != nil {
		return errors.AddContext(err, "version '%v' is not a sem. version", s)
	}

	return nil
}

// Type implements the [pflag.Value] interface.
func (v *Version) Type() string {
	return "ComponentVersion"
}

func (v *Version) UnmarshalText(bytes []byte) error {
	return v.Version.UnmarshalText(bytes)
}

// UnmarshalMapstruct implements the [config.UnmarshalMapstruct] interface.
func (v *Version) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'Version'")
	}

	return v.Set(d)
}
