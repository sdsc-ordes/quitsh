package common

import (
	"fmt"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
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

	panic("Not implemented!")
}

func GetAllBuildTypes() []BuildType {
	return []BuildType{BuildRelease, BuildDebug}
}

// Implement the pflags Value interface.
func (v BuildType) String() string {
	switch v {
	case BuildDebug:
		return BuildDebugName
	case BuildRelease:
		return BuildReleaseName
	}

	panic("Not implemented.")
}

// Implement the pflags Value interface.
func (v *BuildType) Set(s string) (err error) {
	*v, err = NewBuildType(s)

	return
}

// Implement the pflags Value interface.
func (v *BuildType) Type() string {
	return "string"
}

// UnmarshalYAML unmarshals from YAML.
func (v *BuildType) UnmarshalYAML(unmarshal func(any) error) (err error) {
	var s string
	err = unmarshal(&s)
	if err != nil {
		return
	}

	*v, err = NewBuildType(s)

	return
}

// MarshalYAML marshals to YAML.
// Note: needs to be value-receiver to be called!
func (v BuildType) MarshalYAML() (any, error) {
	return v.String(), nil
}

// Implement the [config.UnmarshalMapstruct] interface.
func (v *BuildType) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'BuildType'")
	}

	var err error
	*v, err = NewBuildType(d)

	return err
}
