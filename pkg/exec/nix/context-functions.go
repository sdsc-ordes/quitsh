package nix

import (
	"encoding/json"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
)

type Derivation struct {
	Outputs struct {
		Out string `json:"out"`
	} `json:"outputs"`

	DrvPath string `json:"drvPath"`
}

// BuildInstallable builds the derivation specified by the installable `installable`.
// See [FlakeInstallable].
func (ctx *NixBuildCtx) BuildInstallable(installable string) (*Derivation, error) {
	js, err := ctx.Get("--no-link", "--json", installable)
	if err != nil {
		return nil, err
	}
	var drvs []Derivation

	err = json.Unmarshal([]byte(js), &drvs)
	if err != nil || len(drvs) == 0 {
		return nil, errors.AddContext(err,
			"could not build installable '%v'",
			installable)
	}

	return &drvs[0], nil
}
