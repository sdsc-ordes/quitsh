package imgpkg

import (
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/secret"
)

type (
	Option = func(c *exec.CmdContextBuilder) error
)

// NewCtx returns a new `imgpkg` command context.
func NewCtx(opts ...Option) (*exec.CmdContext, error) {
	b := exec.NewCmdCtxBuilder().
		BaseCmd("imgpkg").
		CredentialFilter(nil)

	for _, o := range opts {
		e := o(&b)
		if e != nil {
			return nil, e
		}
	}

	return b.Build(), nil
}

func WithCredentialsEnv(env secret.CredentialsEnv) Option {
	return func(c *exec.CmdContextBuilder) error {
		creds, e := env.Resolve(false)
		if e != nil {
			return e
		}

		c.Env(
			"IMGPKG_USERNAME="+creds.User(),
			"IMGPKG_TOKEN="+creds.Token(),
		)

		return nil
	}
}
