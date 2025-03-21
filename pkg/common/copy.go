package common

// CopySlice copies a slice.
func CopySlice[T any](s []T) []T {
	res := make([]T, 0, len(s))

	return append(res, s...)
}

// CopySlice copies a slice but giving an initial capacity.
func CopySliceC[T any](s []T, capacity int) []T {
	res := make([]T, 0, capacity)

	return append(res, s...)
}
