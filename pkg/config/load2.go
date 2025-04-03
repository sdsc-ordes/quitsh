package config

import (
	"io"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
	"github.com/spf13/viper"
)

// LoadConfigs reads the config files at location `path` and merges it into
// the existing config in `vIn` (if given).
func loadConfigs[T any, TP Initializable[T]](conf *T, readers ...io.Reader) (err error) {
	v := viper.New()

	// Mapstructure decoder options.
	decOpts := func(c *mapstructure.DecoderConfig) {
		c.TagName = "yaml"
		// Set our special hook to make all types which implement UnmarshalMapstructDecodeHook
		// also deserializable.
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(UnmarshalMapstructDecodeHook)
	}

	// Set defaults for viper.
	var confDefault T
	err = defaults.Set(&confDefault)
	if err != nil {
		return errors.AddContext(err, "Can not set defaults.")
	}
	var defaultsVals map[string]any
	var dO mapstructure.DecoderConfig
	decOpts(&dO)
	dO.Result = &defaultsVals
	dec, err := mapstructure.NewDecoder(&dO)
	if err != nil {
		return err
	}
	err = dec.Decode(&confDefault)
	if err != nil {
		return errors.AddContext(err, "could not decode defaults to set to viper")
	}

	log.Debug("Viper unmarshalling.", "config", conf, "defaults", defaultsVals)
	for k, val := range defaultsVals {
		v.SetDefault(k, val)
	}
	v.SetConfigType("yaml")
	v.SetEnvPrefix("quitsh")
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)
	v.AutomaticEnv()

	for _, f := range readers {
		err = v.MergeConfig(f)
		if err != nil {
			log.ErrorE(err, "Failed to read and merge config file.")

			return err
		}
	}

	log.Info("Viper: ", "settings", v.AllSettings())
	err = v.Unmarshal(&conf, viper.DecoderConfigOption(decOpts))

	if err != nil {
		log.ErrorE(err, "Failed to unmarshal config file.")

		return err
	}

	c := TP(conf)
	err = c.Init()

	if err != nil {
		return err
	}

	return nil
}

// LoadFromReader loads a config file from reader `reader`.
func LoadFromReader2[T any, TP Initializable[T]](
	reader ...io.Reader,
) (conf T, err error) {
	err = LoadFromReaderInto2[T, TP](&conf, reader...)

	return
}

// LoadFromReaderInto loads a config into the type `conf`.
func LoadFromReaderInto2[T any, TP Initializable[T]](
	conf *T,
	reader ...io.Reader,
) error {
	err := loadConfigs[T, TP](conf, reader...)
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
