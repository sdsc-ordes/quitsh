package errors

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorCombinedHandling(t *testing.T) {
	err := errors.New("This is an error.")
	e := Combine(os.ErrNotExist, err)

	require.ErrorIs(t, e, os.ErrNotExist)

	e = Combine(nil, err)
	require.Error(t, e)
	assert.Equal(t, &e, &err, "should return the same")

	e = Combine(nil, nil)
	require.NoError(t, e)

	e = Combine(nil, nil, err)
	require.Error(t, e)
}
