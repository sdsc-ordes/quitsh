package shell

import (
	"slices"
	"strings"

	"deedles.dev/xiter"
)

func escapeQuote(s string) string {
	return strings.ReplaceAll(s, "'", `'\''`)
}

// CmdString returns the joined and single-quoted Shell command.
// This is dumb and should only get used if absolutely necessary.
func CmdToString(args ...string) string {
	quotedArgs := strings.Join(slices.Collect(
		xiter.Map(slices.Values(args),
			func(s string) string { return "'" + escapeQuote(s) + "'" }),
	), " ")

	return quotedArgs
}

// CmdString returns the joined and single-quoted Shell command where
// the first argument is not quoted.
// This is dumb and should only get used if absolutely necessary.
func CmdToStringF(first string, args ...string) string {
	quotedArgs := strings.Join(slices.Collect(
		xiter.Map(slices.Values(args),
			func(s string) string { return "'" + escapeQuote(s) + "'" }),
	), " ")

	return first + " " + quotedArgs
}
