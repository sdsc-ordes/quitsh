package image

import (
	"strings"

	"github.com/sdsc-ordes/quitsh/pkg/errors"

	"github.com/containers/image/v5/docker/reference"
)

type ImageRef reference.Reference
type ImageRefNamed reference.Named
type ImageRefTagged reference.NamedTagged
type ImageRefDigested reference.Digested

// NewRef returns a new image reference (tagged or/and with digest).
func NewRef(
	name string,
	tag string,
	digest string,
) (ImageRef, error) {
	if tag == "" && digest == "" {
		return nil, errors.New("image ref must be either have a tag or a digest")
	}
	n := imageRef{Name: name, Tag: tag, Digest: digest}

	ref, err := reference.Parse(n.String())
	if err != nil {
		return nil, errors.AddContext(err,
			"image reference parsing with name '%s' and tag '%s' and digest '%s' failed",
			name, tag, digest)
	}

	if _, ok := ref.(ImageRefNamed); !ok {
		return nil, errors.New("image name is not an named image '%s'", n.String())
	}

	return ref, nil
}

// NewRefFromString creates an image reference from string `n`.
func NewRefFromString(n string) (ImageRef, error) {
	ref, err := reference.Parse(n)
	if err != nil {
		return nil, err
	}

	if _, ok := ref.(ImageRefNamed); !ok {
		return nil, errors.New("image name is not an named image '%s'", n)
	}

	return ref, nil
}

type imageRef struct {
	Name   string
	Tag    string
	Digest string
}

// String returns the image string.
func (i *imageRef) String() string {
	var sb strings.Builder
	sb.WriteString(i.Name)
	if i.Tag != "" {
		sb.WriteString(":" + i.Tag)
	}
	if i.Digest != "" {
		sb.WriteString("@" + i.Digest)
	}

	return sb.String()
}

// ImageRefField around ImageRef when working with marshaling/unmarshaling.
type ImageRefField struct {
	Ref ImageRef `yaml:",inline"`
}

// UnmarshalYAML unmarshals the image ref.
func (r *ImageRefField) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	err := unmarshal(s)
	if err != nil {
		return err
	}

	r.Ref, err = reference.Parse(s)

	return err
}

// MarshalYAML marshals the image ref. to YAML.
func (r ImageRefField) MarshalYAML() (any, error) {
	return r.Ref.String(), nil
}
