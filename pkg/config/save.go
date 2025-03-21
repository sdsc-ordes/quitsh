package config

import (
	"io"
	"os"

	"github.com/sdsc-ordes/quitsh/pkg/errors"

	"github.com/goccy/go-yaml"
)

// SaveToWriter saves a config to a reader `reader`.
func SaveToWriter[T any](
	conf *T,
	writer io.Writer,
) error {
	return SaveInterfaceToWriter(conf, writer)
}

// SaveInterfaceToWriter saves a config to a reader `reader`.
func SaveInterfaceToWriter(
	conf any,
	writer io.Writer,
) error {
	encoder := yaml.NewEncoder(writer)
	defer encoder.Close()

	return encoder.Encode(conf)
}

// SaveToFile saves a config to a file.
func SaveToFile[T any](
	file string,
	conf *T,
) (err error) {
	return SaveInterfaceToFile(file, conf)
}

// SaveInterfaceToFile saves a config to a file.
func SaveInterfaceToFile(
	file string,
	conf any,
) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return
	}

	defer func() {
		e := f.Close()
		err = errors.Combine(err, e)
	}()

	return SaveInterfaceToWriter(conf, f)
}
