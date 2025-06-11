package debug

import (
	"fmt"

	"github.com/sdsc-ordes/quitsh/pkg/build"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// Assert is a debug functionality and is a no-op in release.
// It will assert that `condition` is `true` and otherwise
// log (formatted not structured) and panic.
func Assert(condition bool, msg string, args ...any) {
	if build.DebugEnabled && !condition {
		log.Debugf(msg, args...)
		panic(fmt.Sprintf("Assert not met: "+msg, args...))
	}
}
