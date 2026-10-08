package common

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/spf13/pflag"
)

type EnvironmentType int

const (
	EnvironmentDev     EnvironmentType = 0
	EnvironmentTesting EnvironmentType = 1
	EnvironmentStaging EnvironmentType = 2
	EnvironmentProd    EnvironmentType = 3
)

func NewEnvironmentType(s string) (EnvironmentType, error) {
	switch s {
	case "dev":
		fallthrough
	case "development":
		return EnvironmentDev, nil
	case "test":
		fallthrough
	case "testing":
		return EnvironmentTesting, nil
	case "stage":
		fallthrough
	case "staging":
		return EnvironmentStaging, nil
	case "prod":
		fallthrough
	case "production":
		return EnvironmentProd, nil
	}

	panic(fmt.Sprintf("EnvironmentType not implemented. '%s'", s))
}

// GetEnvTypes returns all env types.
func GetEnvTypes() []EnvironmentType {
	return []EnvironmentType{
		EnvironmentDev,
		EnvironmentTesting,
		EnvironmentStaging,
		EnvironmentProd,
	}
}

// GetEnvTypesHelp reports some help string for env. types.
func GetEnvTypesHelp() string {
	return fmt.Sprintf(
		"[%s, %s, %s, %s]",
		EnvironmentDev,
		EnvironmentTesting,
		EnvironmentStaging,
		EnvironmentProd,
	)
}

// Interface implementation guard.
var _ pflag.Value = (*EnvironmentType)(nil)

// String implements [pflag.Value].
func (v EnvironmentType) String() string {
	switch v {
	case EnvironmentDev:
		return "development"
	case EnvironmentTesting:
		return "testing"
	case EnvironmentStaging:
		return "staging"
	case EnvironmentProd:
		return "production"
	}

	panic(fmt.Sprintf("EnvironmentType not implemented. %v", int(v)))
}

func (v EnvironmentType) ShortString() string {
	switch v {
	case EnvironmentDev:
		return "dev"
	case EnvironmentTesting:
		return "test"
	case EnvironmentStaging:
		return "stage"
	case EnvironmentProd:
		return "prod"
	}

	panic(fmt.Sprintf("EnvironmentType not implemented. %v", int(v)))
}

// Set implements [pflag.Value].
func (v *EnvironmentType) Set(s string) (err error) {
	*v, err = NewEnvironmentType(s)

	return
}

// Type implements [pflag.Value].
func (v *EnvironmentType) Type() string {
	return "string"
}

// Interface implementation guard.
var _ yaml.InterfaceUnmarshaler = (*EnvironmentType)(nil)

// UnmarshalYAML implements [yaml.InterfaceUnmarshaler].
func (v *EnvironmentType) UnmarshalYAML(unmarshal func(any) error) (err error) {
	var s string

	err = unmarshal(&s)
	if err != nil {
		return
	}

	*v, err = NewEnvironmentType(s)

	return
}

// Interface implementation guard.
var _ yaml.InterfaceMarshaler = (*EnvironmentType)(nil)

// MarshalYAML implements [yaml.InterfaceMarshaler].
// Note: needs to be value-receiver to be called!
func (v EnvironmentType) MarshalYAML() (any, error) {
	return v.String(), nil
}

// Interface implementation guard.
var _ config.UnmarshalerMapstruct = (*EnvironmentType)(nil)

// UnmarshalMapstruct implements [config.UnmarshalerMapstruct].
func (v *EnvironmentType) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'EnvironmentType'")
	}

	var err error
	*v, err = NewEnvironmentType(d)

	return err
}
