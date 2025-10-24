package config

import (
	"io"
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/errors"

	"github.com/goccy/go-yaml"
)

type (
	Initializable[T any] interface {
		Init() error
		*T
	}

	LoadOption func(*opt)

	opt struct {
		noStrict bool
		opts     []yaml.DecodeOption
	}
)

// LoadFromReader loads a config file from reader `reader`.
func LoadFromReader[T any, TP Initializable[T]](
	reader io.Reader,
	opts ...LoadOption,
) (conf T, err error) {
	err = LoadFromReaderInto[T, TP](reader, &conf, opts...)

	return
}

// LoadFromReaderInto loads a config into the type `conf`.
func LoadFromReaderInto[T any, TP Initializable[T]](
	reader io.Reader,
	conf *T,
	opts ...LoadOption,
) error {
	o := opt{}
	o.apply(opts...)

	dec := yaml.NewDecoder(reader, o.opts...)
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
func LoadFromFile[T any, TP Initializable[T]](
	path string,
	opts ...LoadOption,
) (config T, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	config, err = LoadFromReader[T, TP](f, opts...)
	if err != nil {
		err = errors.AddContext(err, "could not load file '%s'", path)
	}

	return
}

func WithLoadNonStrict() LoadOption {
	return func(o *opt) {
		o.noStrict = true
	}
}

func (o *opt) apply(opts ...LoadOption) {
	for i := range opts {
		opts[i](o)
	}

	if !o.noStrict {
		o.opts = append(o.opts, yaml.Strict())
	}
}
