//nolint:forbidigo // intentional
package main

import (
	"fmt"
	"quitsh/tests/component-a/pkg/build"

	"github.com/agnivade/levenshtein"
)

func main() {
	fmt.Printf(
		"Version: %v",
		build.GetBuildVersion(),
	)

	fmt.Printf(
		"Hello World vs. Wello Horld: dist: %v.",
		levenshtein.ComputeDistance("Hello World", "Wello Horld"),
	)
}
