package component

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"

	"github.com/hashicorp/go-version"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/spf13/pflag"
)

type Version struct {
	version.Version
}

// Interface implementation guard.
var _ pflag.Value = (*Version)(nil)

// String implements [pflag.Value].
func (v *Version) String() string {
	return v.Version.String()
}

// Set implements [pflag.Value].
func (v *Version) Set(s string) error {
	err := v.Version.UnmarshalText([]byte(s))
	if err != nil {
		return errors.AddContext(err, "version '%v' is not a sem. version", s)
	}

	return nil
}

// Type implements [pflag.Value].
func (v *Version) Type() string {
	return "ComponentVersion"
}

func (v *Version) UnmarshalText(bytes []byte) error {
	return v.Version.UnmarshalText(bytes)
}

// Interface implementation guard.
var _ config.UnmarshalerMapstruct = (*Version)(nil)

// UnmarshalMapstruct implements [config.UnmarshalerMapstruct].
func (v *Version) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'Version'")
	}

	return v.Set(d)
}
