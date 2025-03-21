//nolint:gochecknoglobals // only private use
package exec

import "strings"

type ArgsFilter func([]string) []string

var defaultCredentialArgs = []string{
	"user",
	"username",
	"password",
	"pass",
	"token",
	"credential",
	"cred",
	"secret",
}

var hiddenDefault = "****hidden****" //nolint:gochecknoglobals // private use.

// DefaultCredentialFilter is the default credential filter which remove secret args from commands.
var DefaultCredentialFilter ArgsFilter = NewCredentialFilter(defaultCredentialArgs)

// NewCredentialFilter returns a filter you can set on `CmdContext`
// for filtering secret arguments you dont want in the log.
func NewCredentialFilter(credArgs []string) ArgsFilter {
	if credArgs == nil {
		credArgs = defaultCredentialArgs
	}

	return func(ss []string) (n []string) {
		n = make([]string, len(ss))
		copy(n, ss)

		filter := func(i, d int) {
			if !strings.HasPrefix(n[i], "-") {
				return
			}

			if strings.Contains(strings.ToLower(n[i]), credArgs[d]) {
				if k := strings.Index(n[i], "="); k >= 0 {
					n[i] = n[i][0:k+1] + hiddenDefault
				} else if i+1 < len(n) {
					n[i+1] = hiddenDefault
				}
			}
		}

		for i := range n {
			for d := range defaultCredentialArgs {
				filter(i, d)
			}
		}

		return
	}
}
