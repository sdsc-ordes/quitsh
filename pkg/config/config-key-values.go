package config

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/goccy/go-yaml"
	"github.com/sdsc-ordes/quitsh/pkg/errors"
	"github.com/sdsc-ordes/quitsh/pkg/log"
)

// ApplyKeyValues applies key,value pairs to a config, e.g. `a.b.c: {"a": 4}`
// would apply `4` in the nested struct fields as in Go code `a.b.c.a = 4`.
// The apply is ordered, successive pairs overwrite/merge with previous
// pairs.
func ApplyKeyValues(keyValues []string, config any) error {
	for i := range keyValues {
		m, err := parseToNestedMap(keyValues[i])
		if err != nil {
			return errors.AddContext(err, "could not parse '%v' to nested map", keyValues[i])
		}

		// Mapstructure decoder options.
		var dO mapstructure.DecoderConfig
		dO.TagName = `yaml`
		dO.SquashTagOption = `inline`
		dO.ErrorUnused = true
		dO.WeaklyTypedInput = true

		// Set our special hook to make all types
		// which implement UnmarshalMapstructDecodeHook
		// also deserializable.
		dO.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			UnmarshalMapstructDecodeHook)

		dO.Result = config

		dec, err := mapstructure.NewDecoder(&dO)
		if err != nil {
			return errors.AddContext(
				err,
				"could not create mapstructure decoder for '%v'",
				keyValues[i],
			)
		}

		err = dec.Decode(m)
		if err != nil {
			return errors.AddContext(err, "could not decode '%v' and apply to config", keyValues[i])
		}
	}

	return nil
}

func parseToNestedMap(kv string) (map[string]any, error) {
	res := make(map[string]any)

	s := strings.SplitN(kv, ":", 2) //nolint:mnd
	if len(s) != 2 || s[1] == "" {
		return nil, errors.New(
			"key value pair '%v' is not of the form '<path>:<yaml-data>', e.g. `a.b.c.D.e: true`",
			kv,
		)
	}

	keys := strings.Split(s[0], ".")
	if len(keys) == 0 {
		return nil, errors.New("must have at least one key in '%v'", s[0])
	}

	log.Trace("YAML unmarshal '%s'", s[1])
	var value any
	err := yaml.Unmarshal([]byte(strings.TrimSpace(s[1])), &value)
	if err != nil {
		return nil, errors.AddContext(err, "could not unmrashal YAML '%v'", s[1])
	}

	// Build recursive map for mapstructure decoding until last value.
	curr := &res
	for _, key := range keys[0:max(0, len(keys)-1)] {
		if key == "" {
			return nil, errors.New("empty key in '%v' not allowed", s[0])
		}

		m := make(map[string]any)
		(*curr)[key] = &m
		curr = &m
	}

	(*curr)[keys[len(keys)-1]] = value

	return res, nil
}
