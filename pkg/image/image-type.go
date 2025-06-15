package image

import (
	"errors"
	"fmt"
)

type Type int

const (
	// If you change this here -> adjust the `New*` functions.
	ImageService         Type = 0
	ImageDBMigration     Type = 1
	ImageServiceName          = "service"
	ImageDBMigrationName      = "dbmigration"
)

func NewType(s string) (Type, error) {
	switch s {
	case ImageServiceName:
		return ImageService, nil
	case ImageDBMigrationName:
		return ImageDBMigration, nil
	}

	return 0, fmt.Errorf("wrong build type '%s'", s)
}

// GetImageTypesHelp reports some help string for image types.
func GetImageTypesHelp() string {
	return fmt.Sprintf("[%s, %s]", ImageServiceName, ImageDBMigrationName)
}

// GetAllImageTypes returns all possible image types.
func GetAllImageTypes() []Type {
	return []Type{ImageService, ImageDBMigration}
}

// Implement the pflags Value interface.
func (v Type) String() string {
	switch v {
	case ImageService:
		return ImageServiceName
	case ImageDBMigration:
		return ImageDBMigrationName
	}

	panic("Not implemented.")
}

// Implement the pflags Value interface.
func (v *Type) Set(s string) (err error) {
	*v, err = NewType(s)

	return
}

// Implement the pflags Value interface.
func (v *Type) Type() string {
	return v.String()
}

// UnmarshalYAML unmarshals from YAML.
func (v *Type) UnmarshalYAML(unmarshal func(any) error) (err error) {
	var s string
	err = unmarshal(&s)
	if err != nil {
		return
	}

	*v, err = NewType(s)

	return
}

// MarshalYAML marshals to YAML.
// Note: needs to be value-receiver to be called!
func (v Type) MarshalYAML() (any, error) {
	return v.String(), nil
}

// Implement the [config.UnmarshalMapstruct] interface.
func (v *Type) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'EnvironmentType'")
	}

	var err error
	*v, err = NewType(d)

	return err
}
