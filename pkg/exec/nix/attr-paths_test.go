package nix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlakeInstallable(t *testing.T) {
	assert.Equal(t, "./#banana", FlakeInstallable("", "banana"))
	assert.Equal(t, "./test/bla#banana", FlakeInstallable("test/bla", "banana"))
	assert.Equal(t, "./test/bla#banana", FlakeInstallable("./test/bla", "banana"))
	assert.Equal(t, "/test/bla#banana", FlakeInstallable("/test/bla", "banana"))
}
