package skopeo

import (
	"fmt"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/sdsc-ordes/quitsh/pkg/secret"
)

type (
	Context struct {
		*exec.CmdContext

		builder contextBuilder
	}

	contextBuilder struct {
		exec.CmdContextBuilder
		useTLS bool
	}

	Option = func(c *contextBuilder)
)

// NewCtx returns a new `skopeo` command context.
func NewCtx(opts ...Option) Context {
	b := contextBuilder{exec.NewCmdCtxBuilder().
		BaseCmd("skopeo").
		CredentialFilter(nil), true}

	for _, o := range opts {
		o(&b)
	}

	return Context{CmdContext: b.Build(), builder: b}
}

// WithEnableTLS enables TLS (https) on skopeo.
func WithEnableTLS(enable bool) Option {
	return func(c *contextBuilder) {
		c.useTLS = enable
	}
}

// Login logs into the registry.
func (s Context) Login(creds secret.Credentials, domain string) (logout func() error, err error) {
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

// InspectCtx returns an inspect ctx.
func (s Context) InspectCtx() Context {
	b := s.builder.Clone()

	return Context{
		CmdContext: addTLS(b.BaseArgs("inspect"), s.builder.useTLS, false).Build(),
		builder:    s.builder,
	}
}

// CopyCtx returns an copy ctx.
func (s Context) CopyCtx() Context {
	b := s.builder.Clone()

	return Context{
		CmdContext: addTLS(b.BaseArgs("copy"), s.builder.useTLS, true).Build(),
		builder:    s.builder,
	}
}

func addTLS(b exec.CmdContextBuilder, useTLS bool, forCopy bool) exec.CmdContextBuilder {
	if forCopy {
		return b.BaseArgs(
			fmt.Sprintf("--src-tls-verify=%v", useTLS),
			fmt.Sprintf("--dest-tls-verify=%v", useTLS))
	} else {
		return b.BaseArgs(
			fmt.Sprintf("--tls-verify=%v", useTLS))
	}
}
