package secret

import (
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/exec/env"
)

type CredentialsEnv struct {
	UserEnv  string `yaml:"userEnv"`
	TokenEnv string `yaml:"tokenEnv"`
}

type Credentials struct {
	user  RedactedString
	token RedactedString
}

// NewCredentials returns new credentials from env. variables.
func NewCredentials(envs CredentialsEnv) (c Credentials, err error) {
	return envs.Resolve(false)
}

// NewCredentialsTokenOnly creates credentials only for the token (user is `quitsh`).
func NewCredentialsTokenOnly(tokenEnv string) (c Credentials, err error) {
	env := CredentialsEnv{TokenEnv: tokenEnv}

	return env.Resolve(true)
}

// ResolveFrom resolves credentials from the given environment.
func (e *CredentialsEnv) ResolveFrom(tokenOnly bool, environ []string) (c Credentials, err error) {
	all := []string{e.TokenEnv}
	if !tokenOnly {
		all = []string{e.UserEnv, e.TokenEnv}
	}

	err = env.AssertProperEnvKey(all...)
	if err != nil {
		return
	}

	l := env.EnvList(environ).FindAll(all...)
	err = l.AssertNotEmpty()
	if err != nil {
		return
	}

	user := "quitsh"
	if !tokenOnly {
		user = l[e.UserEnv].Value
	}

	return Credentials{
		user:  RedactedString(user),
		token: RedactedString(l[e.TokenEnv].Value)}, nil
}

// Resolve all credential env variables from the `os` environment.
func (e *CredentialsEnv) Resolve(tokenOnly bool) (c Credentials, err error) {
	return e.ResolveFrom(tokenOnly, os.Environ())
}

func (c *Credentials) User() string {
	return string(c.user)
}

func (c *Credentials) Token() string {
	return string(c.token)
}
