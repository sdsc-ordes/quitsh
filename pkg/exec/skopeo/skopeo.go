package skopeo

import (
	"fmt"

	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type (
	Context struct {
		*exec.CmdContext
	}

	Option = func(c exec.CmdContextBuilder)
)

// NewCtx returns a new `skopeo` command context.
func NewCtx(opts ...Option) Context {
	b := exec.NewCmdCtxBuilder().
		BaseCmd("skopeo").
		CredentialFilter(nil)

	for _, o := range opts {
		o(b)
	}

	return Context{b.Build()}
}

// WithEnableTLS enables TLS (https) on skopeo.
func WithEnableTLS(enable bool) Option {
	return func(c exec.CmdContextBuilder) {
		c.BaseArgs(fmt.Sprintf("--tls-verify=%v", enable))
	}
}

// Login logs into the registry.
func (s Context) Login(creds common.Credentials, domain string) (logout func() error, err error) {
	log.Info("Login to registry.", "domain", domain)

	err = s.Check("login",
		"--username",
		creds.User(),
		"--password",
		creds.Token(),
		domain)

	if err != nil {
		return nil, err
	}

	logout = func() error {
		e := s.Check("logout", domain)
		if e != nil {
			return errors.AddContext(e, "could not logout from skopeo")
		}

		return nil
	}

	return
}
