package image

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Name   string
	Tag    string
	Digest string

	Valid bool
}

func TestImageName(t *testing.T) {
	tests := []TestData{
		{Name: "bla.com/a/b/c/d", Tag: "1.2.3", Valid: true},
		{
			Name:   "bla.com/a/b/c/d",
			Tag:    "",
			Digest: "sha256:7ca7f383f2beb9a8fe876b2fb7601e12370f748f899ea9b95620e1d7b08f000b",
			Valid:  true,
		},
		{Name: "bla.com/a/b/c/d", Tag: "", Digest: "", Valid: false},
		{Name: "bla.com/a/b/c/d", Tag: "1.0.2+asdf", Digest: "", Valid: false},
		{Name: "", Tag: "1.0.2+asdf", Digest: "", Valid: false},
	}

	for _, test := range tests {
		ref, err := NewRef(test.Name, test.Tag, test.Digest)
		if test.Valid {
			assert.NoError(t, err) //nolint:testifylint // intentional.

			if r, ok := ref.(ImageRefTagged); ok {
				assert.Equal(t, test.Tag, r.Tag())
			}

			if r, ok := ref.(ImageRefNamed); ok {
				assert.Equal(t, test.Name, r.Name())
			}
		} else {
			assert.Error(t, err)
		}
	}
}
