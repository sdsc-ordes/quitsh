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
	err = env.AssertProperEnvKey(envs.UserEnv, envs.TokenEnv)
	if err != nil {
		return
	}

	l := env.EnvList(os.Environ()).FindAll(envs.UserEnv, envs.TokenEnv)
	err = l.AssertNotEmpty()
	if err != nil {
		return
	}

	return Credentials{
		user:  l[envs.UserEnv].Value,
		token: l[envs.TokenEnv].Value}, nil
}

// NewCredentialsTokenOnly creates credentials only for the token (user is `quitsh`).
func NewCredentialsTokenOnly(tokenEnv string) (c Credentials, err error) {
	err = env.AssertProperEnvKey(tokenEnv)
	if err != nil {
		return
	}

	l := env.EnvList(os.Environ()).FindAll(tokenEnv)
	err = l.AssertNotEmpty()
	if err != nil {
		return
	}

	return Credentials{
		user:  "quitsh",
		token: l[tokenEnv].Value}, nil
}

func (c Credentials) User() string {
	return c.user
}

func (c Credentials) Token() string {
	return c.token
}
