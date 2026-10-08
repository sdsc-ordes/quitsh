package image

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/sdsc-ordes/quitsh/pkg/config"
	"github.com/spf13/pflag"
)

type Type int

const (
	// If you change this here -> adjust the `New*` functions.
	ImageService     Type = 0 // A service image.
	ImageDBMigration Type = 1 // A DB migration image.
	ImageBundle      Type = 2 // A manifest bundle from `imgpkg` or similar.
	ImageData        Type = 3 // A data image with only files.

	ImageServiceName     = "service"
	ImageDBMigrationName = "dbmigration"
	ImageBundleName      = "bundle"
	ImageDataName        = "data"
)

func NewType(s string) (Type, error) {
	switch s {
	case ImageServiceName:
		return ImageService, nil
	case ImageDBMigrationName:
		return ImageDBMigration, nil
	case ImageBundleName:
		return ImageBundle, nil
	case ImageDataName:
		return ImageData, nil
	}

	return 0, fmt.Errorf("wrong build type '%s'", s)
}

// GetImageTypesHelp reports some help string for image types.
func GetImageTypesHelp() string {
	return fmt.Sprintf(
		"[%s, %s, %s, %s]",
		ImageServiceName,
		ImageDBMigrationName,
		ImageBundle,
		ImageData,
	)
}

// GetAllImageTypes returns all possible image types.
func GetAllImageTypes() []Type {
	return []Type{ImageService, ImageDBMigration, ImageBundle, ImageData}
}

// Interface implementation guard.
var _ pflag.Value = (*Type)(nil)

// String implements [pflag.Value].
func (v Type) String() string {
	switch v {
	case ImageService:
		return ImageServiceName
	case ImageDBMigration:
		return ImageDBMigrationName
	case ImageBundle:
		return ImageBundleName
	case ImageData:
		return ImageDataName
	}

	panic("Not implemented.")
}

// Set implements [pflag.Value].
func (v *Type) Set(s string) (err error) {
	*v, err = NewType(s)

	return
}

// Type implements [pflag.Value].
func (v *Type) Type() string {
	return v.String()
}

// Interface implementation guard.
var _ yaml.InterfaceUnmarshaler = (*Type)(nil)

// UnmarshalYAML implements [yaml.InterfaceUnmarshaler].
func (v *Type) UnmarshalYAML(unmarshal func(any) error) (err error) {
	var s string
	err = unmarshal(&s)
	if err != nil {
		return
	}

	*v, err = NewType(s)

	return
}

// Interface implementation guard.
var _ yaml.InterfaceMarshaler = (*Type)(nil)

// MarshalYAML implements [yaml.InterfaceMarshaler].
// Note: needs to be value-receiver to be called!
func (v Type) MarshalYAML() (any, error) {
	return v.String(), nil
}

// Interface implementation guard.
var _ config.UnmarshalerMapstruct = (*Type)(nil)

// UnmarshalMapstruct implements [config.UnmarshalerMapstruct].
func (v *Type) UnmarshalMapstruct(data any) error {
	d, ok := data.(string)
	if !ok {
		return errors.New("can only unmarshal from 'string' into 'EnvironmentType'")
	}

	var err error
	*v, err = NewType(d)

	return err
}
