package image

import (
	"errors"
	"fmt"
)

type Type int

const (
	// If you change this here -> adjust the `New*` functions.
	ImageService         Type = 0 // A service migration image.
	ImageDBMigration     Type = 1 // A DB migration image.
	ImageBundle          Type = 2 // A manifest bundle from `imgpkg` or similar.
	ImageServiceName          = "service"
	ImageDBMigrationName      = "dbmigration"
	ImageBundleName           = "bundle"
)

func NewType(s string) (Type, error) {
	switch s {
	case ImageServiceName:
		return ImageService, nil
	case ImageDBMigrationName:
		return ImageDBMigration, nil
	case ImageBundleName:
		return ImageBundle, nil
	}

	return 0, fmt.Errorf("wrong build type '%s'", s)
}

// GetImageTypesHelp reports some help string for image types.
func GetImageTypesHelp() string {
	return fmt.Sprintf("[%s, %s, %s]", ImageServiceName, ImageDBMigrationName, ImageBundle)
}

// GetAllImageTypes returns all possible image types.
func GetAllImageTypes() []Type {
	return []Type{ImageService, ImageDBMigration, ImageBundle}
}

// String implements the interface [pflags.Value].
func (v Type) String() string {
	switch v {
	case ImageService:
		return ImageServiceName
	case ImageDBMigration:
		return ImageDBMigrationName
	case ImageBundle:
		return ImageBundleName
	}

	panic("Not implemented.")
}

// Set implements the interface [pflags.Value].
func (v *Type) Set(s string) (err error) {
	*v, err = NewType(s)

	return
}

// Type implements the interface [pflags.Value].
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

// UnmarshalMapstruct implements the [config.UnmarshalMapstruct] interface.
func (v *Type) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'EnvironmentType'")
	}

	var err error
	*v, err = NewType(d)

	return err
}
