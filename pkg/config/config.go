package config

type IConfig interface {
	Clone() IConfig
}
