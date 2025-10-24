package config

// EnvExpander is an interface which `[IConfig]` passed to
// quitsh can implement to replace env. variables in
// certain settings if needed.
type EnvExpander interface {
	ExpandEnv() error
}
