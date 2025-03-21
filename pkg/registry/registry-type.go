package registry

import (
	"fmt"
)

type RegistryType int

const (
	// If you change this here -> adjust the `New*` functions.
	RegistryTemp        RegistryType = 0
	RegistryRelease     RegistryType = 1
	RegistryTempName                 = "temporary"
	RegistryReleaseName              = "release"
)

func NewRegistryType(s string) (RegistryType, error) {
	switch s {
	case RegistryReleaseName:
		return RegistryRelease, nil
	case RegistryTempName:
		return RegistryTemp, nil
	}

	return 0, fmt.Errorf("wrong build type '%s'", s)
}

// Implement the pflags Value interface.
func (v RegistryType) String() string {
	switch v {
	case RegistryRelease:
		return RegistryReleaseName
	case RegistryTemp:
		return RegistryTempName
	}

	panic("Not implemented.")
}

// Implement the pflags Value interface.
func (v *RegistryType) Set(s string) (err error) {
	*v, err = NewRegistryType(s)

	return
}

// GetAllRegistryTypes returns all registry types.
func GetAllRegistryTypes() []RegistryType {
	return []RegistryType{RegistryTemp, RegistryRelease}
}

// Implement the pflags Value interface.
func (v *RegistryType) Type() string {
	return "string"
}

// UnmarshalYAML unmarshals from YAML.
func (v *RegistryType) UnmarshalYAML(unmarshal func(any) error) (err error) {
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
func (v RegistryType) MarshalYAML() (any, error) {
	return v.String(), nil
}
