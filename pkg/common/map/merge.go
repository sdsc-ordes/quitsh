package maps

// Merge copies only entries from `src` to `dst` and returns the existing keys already
// in `dest`.
func Merge[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) []K {
	var existing []K
	if src == nil {
		return nil
	}

	for k, v := range src {
		if _, exists := dst[k]; exists {
			existing = append(existing, k)
		}
		dst[k] = v
	}

	return existing
}
