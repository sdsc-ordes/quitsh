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
) (err error) {
	encoder := yaml.NewEncoder(writer)
	defer func() {
		err = errors.Combine(err, encoder.Close())
	}()

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
	var f *os.File
	if file != "-" {
		f, err = os.Create(file)
		if err != nil {
			return
		}

		defer func() {
			err = errors.Combine(err, f.Close())
		}()
	} else {
		f = os.Stdout
	}

	return SaveInterfaceToWriter(conf, f)
}
