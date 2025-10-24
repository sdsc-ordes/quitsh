package config

import (
	"io"
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/errors"

	"github.com/goccy/go-yaml"
)

type Initializable[T any] interface {
	Init() error
	*T
}

// LoadFromReader loads a config file from reader `reader`.
func LoadFromReader[T any, TP Initializable[T]](
	reader io.Reader,
) (conf T, err error) {
	err = LoadFromReaderInto[T, TP](reader, &conf)

	return
}

// LoadFromReaderInto loads a config into the type `conf`.
func LoadFromReaderInto[T any, TP Initializable[T]](
	reader io.Reader,
	conf *T,
) error {
	dec := yaml.NewDecoder(reader, yaml.Strict())
	err := dec.Decode(conf)
	if err != nil {
		return err
	}

	c := TP(conf)
	err = c.Init()

	if err != nil {
		return err
	}

	return nil
}

// LoadFromFile loads a config file from `path`.
func LoadFromFile[T any, TP Initializable[T]](path string) (config T, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	config, err = LoadFromReader[T, TP](f)
	if err != nil {
		err = errors.AddContext(err, "could not load file '%s'", path)
	}

	return
}
