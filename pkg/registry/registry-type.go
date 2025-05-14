package registry

import (
	"fmt"
)

type Type int

const (
	// If you change this here -> adjust the `New*` functions.
	RegistryTemp         Type = 0
	RegistryRelease      Type = 1
	RegistryTiltRegistry Type = 2

	RegistryTempName    = "temporary"
	RegistryReleaseName = "release"

	RegistryTiltName = "tilt"
)

func NewRegistryType(s string) (Type, error) {
	switch s {
	case RegistryReleaseName:
		return RegistryRelease, nil
	case RegistryTempName:
		return RegistryTemp, nil
	case RegistryTiltName:
		return RegistryTiltRegistry, nil
	}

	return 0, fmt.Errorf("wrong registry type '%s'", s)
}

// Implement the pflags Value interface.
func (v Type) String() string {
	switch v {
	case RegistryRelease:
		return RegistryReleaseName
	case RegistryTemp:
		return RegistryTempName
	case RegistryTiltRegistry:
		return RegistryTiltName
	}

	panic("Not implemented.")
}

// Implement the pflags Value interface.
func (v *Type) Set(s string) (err error) {
	*v, err = NewRegistryType(s)

	return
}

// GetAllRegistryTypes returns all registry types.
func GetAllRegistryTypes() []Type {
	return []Type{RegistryTemp, RegistryRelease, RegistryTiltRegistry}
}

// Implement the pflags Value interface.
func (v *Type) Type() string {
	return "string"
}

// UnmarshalYAML unmarshals from YAML.
func (v *Type) UnmarshalYAML(unmarshal func(any) error) (err error) {
	var s string
	err = unmarshal(&s)
	if err != nil {
		return
	}

	*v, err = NewRegistryType(s)

	return
}

// MarshalYAML marshals to YAML.
// Note: needs to be value-receiver to be called!
func (v Type) MarshalYAML() (any, error) {
	return v.String(), nil
}
