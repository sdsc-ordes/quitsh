package concurrent

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap_AllSuccess(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	run := func(_ int) error {
		return nil
	}

	run2 := func(_, _ int) error {
		return nil
	}

	err := Map(slices.Values(items), run)
	require.NoError(t, err)

	err = Map2(slices.All(items), run2)
	require.NoError(t, err)
}

func TestMap_WithErrors(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	run := func(p int) error {
		if p%2 == 0 {
			return errors.New("fail for even")
		}

		return nil
	}

	run2 := func(_, p int) error { return run(p) }

	err := Map(slices.Values(items), run)
	require.ErrorContains(t, err, "fail for even")

	err = Map2(slices.All(items), run2)
	require.ErrorContains(t, err, "fail for even")
}
