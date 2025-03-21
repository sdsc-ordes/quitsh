package dag

// InputChanges represents the tracking state of the input change set.
type InputChanges struct {
	// If this input change set contains changes.
	Changed bool

	// Paths which have changed (this list may not represent all paths)
	// because of early returns.
	Changes []string
}
