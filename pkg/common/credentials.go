package common

import (
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/exec/env"
)

type CredentialsEnv struct {
	UserEnv  string `yaml:"userEnv"`
	TokenEnv string `yaml:"tokenEnv"`
}

type Credentials struct {
	user  string `yaml:"-"`
	token string `yaml:"-"`
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

// Resolve all credential env variables.
func (e *CredentialsEnv) Resolve(tokenOnly bool) (c Credentials, err error) {
	all := []string{e.TokenEnv}
	if !tokenOnly {
		all = []string{e.UserEnv, e.TokenEnv}
	}

	err = env.AssertProperEnvKey(all...)
	if err != nil {
		return
	}

	l := env.EnvList(os.Environ()).FindAll(all...)
	err = l.AssertNotEmpty()
	if err != nil {
		return
	}

	user := "quitsh"
	if !tokenOnly {
		user = l[e.UserEnv].Value
	}

	return Credentials{
		user:  user,
		token: l[e.TokenEnv].Value}, nil
}

func (c *Credentials) User() string {
	return c.user
}

func (c *Credentials) Token() string {
	return c.token
}
