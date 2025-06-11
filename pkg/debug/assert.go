package debug

import (
	"fmt"

	"github.com/sdsc-ordes/quitsh/pkg/build"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// Assert is a debug functionality and is a no-op in release.
// It will assert that `condition` is `true` and otherwise
// log (structured) and panic.
func Assert(condition bool, msg string, args ...any) {
	if build.DebugEnabled && !condition {
		log.Error(msg, args...)
		panic("Assert not met: '" + msg + "' -> See in debug log above.")
	}
}

// Assertf same as [Assert] but with formatting.
func Assertf(condition bool, msg string, args ...any) {
	if build.DebugEnabled && !condition {
		log.Errorf(msg, args...)
		panic(fmt.Sprintf("Assert not met: "+msg, args...))
	}
}
