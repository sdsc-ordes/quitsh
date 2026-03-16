package config

type IConfig interface {
	Clone() IConfig
	Validate() error
}
