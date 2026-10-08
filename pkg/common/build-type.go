package common

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/spf13/pflag"
)

type BuildType int

const (
	// If you change this here -> adjust the `New*` functions.
	BuildDebug       BuildType = 0
	BuildRelease     BuildType = 1
	BuildReleaseName           = "release"
	BuildDebugName             = "debug"
)

func NewBuildType(s string) (BuildType, error) {
	switch s {
	case BuildDebugName:
		return BuildDebug, nil
	case BuildReleaseName:
		return BuildRelease, nil
	}

	return 0, fmt.Errorf("wrong build type '%s'", s)
}

func NewBuildTypeFromEnv(e EnvironmentType) BuildType {
	switch e {
	case EnvironmentDev:
		return BuildDebug
	case EnvironmentTesting:
		fallthrough
	case EnvironmentStaging:
		fallthrough
	case EnvironmentProd:
		return BuildRelease
	}

	panic(fmt.Sprintf("BuildType not implemented. '%s'", e))
}

func GetAllBuildTypes() []BuildType {
	return []BuildType{BuildRelease, BuildDebug}
}

// Interface implementation guard.
var _ pflag.Value = (*BuildType)(nil)

// String implements [pflag.Value].
func (v BuildType) String() string {
	switch v {
	case BuildDebug:
		return BuildDebugName
	case BuildRelease:
		return BuildReleaseName
	}

	panic(fmt.Sprintf("BuildType not implemented. '%v'", int(v)))
}

// Set implements [pflag.Value].
func (v *BuildType) Set(s string) (err error) {
	*v, err = NewBuildType(s)

	return
}

// Type implements [pflag.Value].
func (v *BuildType) Type() string {
	return "string"
}

// Interface implementation guard.
var _ yaml.InterfaceUnmarshaler = (*BuildType)(nil)

// UnmarshalYAML implements [yaml.InterfaceUnmarshaler].
func (v *BuildType) UnmarshalYAML(unmarshal func(any) error) (err error) {
	var s string
	err = unmarshal(&s)
	if err != nil {
		return
	}

	*v, err = NewBuildType(s)

	return
}

// Interface implementation guard.
var _ yaml.InterfaceMarshaler = (*BuildType)(nil)

// MarshalYAML implements [yaml.InterfaceMarshaler].
// Note: needs to be value-receiver to be called!
func (v BuildType) MarshalYAML() (any, error) {
	return v.String(), nil
}

// Interface implementation guard.
var _ config.UnmarshalerMapstruct = (*BuildType)(nil)

// UnmarshalMapstruct implements [config.UnmarshalerMapstruct].
func (v *BuildType) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'BuildType'")
	}

	var err error
	*v, err = NewBuildType(d)

	return err
}
