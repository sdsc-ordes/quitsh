package set

import "log/slog"

// Interface implementation guard.
var _ slog.LogValuer = (*Unordered[string])(nil)

// LogValue implements [slog.LogValuer].
func (s Unordered[T]) LogValue() slog.Value {
	// Convert set keys to a slice for logging
	var values []T
	for key := range s.set {
		values = append(values, key)
	}

	// Return the slice as a slog.Value
	return slog.AnyValue(values)
}
