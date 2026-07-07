package common

import "slices"

// CopySlice copies a slice.
// The new slice stays `nil` if `s` is nil.
func CopySlice[T any](s []T) []T {
	return slices.Clone(s)
}

// CopySliceC copies a slice but giving an initial capacity.
// The new slice stays `nil` if `s` is nil.
func CopySliceC[T any](s []T, capacity int) []T {
	if s == nil {
		return nil
	}

	res := make([]T, 0, capacity)

	return append(res, s...)
}
