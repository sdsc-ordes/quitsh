package target

import "strings"

const NamespaceSeparator = "::"

type ID string

// Name returns the name part (second part) after "::".
func (i *ID) Name() string {
	s := (string)(*i)
	if idx := strings.Index(s, NamespaceSeparator); idx > 0 {
		return s[idx+2:]
	}

	return s
}

// Namespace returns the namespace part (first part) before "::"
// if it exists.
func (i *ID) Namespace() (n string, exists bool) {
	s := (string)(*i)
	if idx := strings.Index(s, NamespaceSeparator); idx > 0 {
		return s[:idx], true
	}

	return "", false
}

// String returns the string of the ID.
func (i *ID) String() string {
	return (string)(*i)
}
