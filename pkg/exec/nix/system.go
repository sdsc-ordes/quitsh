package nix

import (
	"github.com/sdsc-ordes/quitsh/pkg/errors"
)

// CurrentSystem reports the current system.
func CurrentSystem() (currentSystem string, err error) {
	nixctx := NewCtxBuilder().Build()
	currentSystem, err = nixctx.Get(
		"eval",
		"--impure",
		"--raw",
		"--expr",
		"builtins.currentSystem",
	)

	if err != nil {
		err = errors.AddContext(
			err,
			"could not get current system from Nix.",
		)
	}

	return
}
