package strs

import (
	"strings"
)

// SplitLines splits a string into an array of strings.
func SplitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

// Indent indents a string `s` with indentation string `ind`.
// Replaces all `\r\n` to `\n`.
func Indent(s string, ind string) string {
	s = ind + strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", "\n"+ind)

	return s
}

// SplitAndTrim splits a string with delimiter `del` and trims all
// spaces from each value.
func SplitAndTrim(s string, del string) []string {
	l := strings.Split(s, del)
	ls := make([]string, 0, len(l))

	for i := range l {
		ls = append(ls, strings.TrimSpace(l[i]))
	}

	return ls
}
