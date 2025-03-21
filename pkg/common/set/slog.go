package set

import "log/slog"

// LogValue implements slog.LogValuer for structured logging.
func (s Unordered[T]) LogValue() slog.Value {
	// Convert set keys to a slice for logging
	var values []T
	for key := range s.set {
		values = append(values, key)
	}

	// Return the slice as a slog.Value
	return slog.AnyValue(values)
}
