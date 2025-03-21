package skopeo

import (
	"github.com/sdsc-ordes/quitsh/pkg/common"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/exec"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

type Context struct {
	*exec.CmdContext
}

// NewCtx returns a new `skopeo` command context.
func NewCtx() Context {
	return Context{
		exec.NewCmdCtxBuilder().
			BaseCmd("skopeo").
			CredentialFilter(nil).Build()}
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
