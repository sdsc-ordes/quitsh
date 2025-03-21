package image

import (
	"fmt"
)

type ImageType int

const (
	// If you change this here -> adjust the `New*` functions.
	ImageService         ImageType = 0
	ImageDBMigration     ImageType = 1
	ImageServiceName               = "service"
	ImageDBMigrationName           = "dbmigration"
)

func NewImageType(s string) (ImageType, error) {
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
func GetAllImageTypes() []ImageType {
	return []ImageType{ImageService, ImageDBMigration}
}

// Implement the pflags Value interface.
func (v ImageType) String() string {
	switch v {
	case ImageService:
		return ImageServiceName
	case ImageDBMigration:
		return ImageDBMigrationName
	}

	panic("Not implemented.")
}

// Implement the pflags Value interface.
func (v *ImageType) Set(s string) (err error) {
	*v, err = NewImageType(s)

	return
}

// Implement the pflags Value interface.
func (v *ImageType) Type() string {
	return v.String()
}

// UnmarshalYAML unmarshals from YAML.
func (v *ImageType) UnmarshalYAML(unmarshal func(any) error) (err error) {
	var s string
	err = unmarshal(&s)
	if err != nil {
		return
	}

	*v, err = NewImageType(s)

	return
}

// MarshalYAML marshals to YAML.
// Note: needs to be value-receiver to be called!
func (v ImageType) MarshalYAML() (any, error) {
	return v.String(), nil
}
